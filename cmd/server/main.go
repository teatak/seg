package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/teatak/seg/pkg/dict"
	"github.com/teatak/seg/pkg/learn"
	"github.com/teatak/seg/pkg/seg"
)

// Engine 分词引擎
type Engine struct {
	prodSegmenter *seg.Segmenter // P roduction segmenter (useStaging=false)
	evalSegmenter *seg.Segmenter // Evaluation segmenter (useStaging=true)
	dictionary    *dict.Dictionary
	feedbackStore *learn.FeedbackStore
	miner         *learn.Miner
	updater       *learn.Updater
	requestLogger *learn.RequestLogger
}

// Config 配置
type Config struct {
	Port        int
	DataDir     string
	BaseDict    string
	UserDict    string
	StagingDict string
	FeedbackDB  string
	RequestLog  string
}

var engine *Engine
var config Config

func main() {
	// 解析命令行参数
	flag.IntVar(&config.Port, "port", 8080, "HTTP server port")
	flag.StringVar(&config.DataDir, "data", "./data", "Data directory")
	flag.Parse()

	// 初始化路径
	config.BaseDict = filepath.Join(config.DataDir, "dict/base.txt")
	config.UserDict = filepath.Join(config.DataDir, "dict/user.txt")
	config.StagingDict = filepath.Join(config.DataDir, "dict/staging.txt")
	config.FeedbackDB = filepath.Join(config.DataDir, "feedback.json")
	config.RequestLog = filepath.Join(config.DataDir, "requests.json")

	// 确保目录存在
	os.MkdirAll(filepath.Dir(config.UserDict), 0755)

	// 初始化引擎
	var err error
	engine, err = NewEngine(config)
	if err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}

	// 设置路由
	http.HandleFunc("/api/segment", handleSegment)
	http.HandleFunc("/api/segment/search", handleSegmentSearch)
	http.HandleFunc("/api/feedback", handleFeedback)
	http.HandleFunc("/api/words", handleWords)
	http.HandleFunc("/api/words/list", handleWordList)
	http.HandleFunc("/api/words/search", handleWordSearchAPI)
	http.HandleFunc("/api/words/", handleWordOps)
	http.HandleFunc("/api/dict/merge", handleMergeDict) // New endpoint
	http.HandleFunc("/api/learn", handleLearn)
	http.HandleFunc("/api/learn/requests", handleLearnFromRequests)
	http.HandleFunc("/api/stats", handleStats)
	http.HandleFunc("/", handleIndex)

	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// NewEngine 创建引擎
func NewEngine(cfg Config) (*Engine, error) {
	// 创建词典 (Init with staging dict)
	dictionary := dict.NewDictionary(cfg.BaseDict, cfg.UserDict, cfg.StagingDict)
	if err := dictionary.Load(); err != nil {
		return nil, fmt.Errorf("failed to load dictionary: %w", err)
	}

	// 创建 HMM 模型
	hmm := seg.NewDefaultHMM()

	// 创建分词器 (Dual instances)
	prodSegmenter := seg.NewSegmenter(dictionary, hmm, false)
	evalSegmenter := seg.NewSegmenter(dictionary, hmm, true)

	// 创建反馈存储
	feedbackStore := learn.NewFeedbackStore(cfg.FeedbackDB)
	feedbackStore.Load()

	// 创建新词发现器
	miner := learn.NewMiner()

	// 创建更新器
	updater := learn.NewUpdater(dictionary, feedbackStore, miner)

	// 创建请求记录器
	requestLogger := learn.NewRequestLogger(cfg.RequestLog)
	requestLogger.Load()

	return &Engine{
		prodSegmenter: prodSegmenter,
		evalSegmenter: evalSegmenter,
		dictionary:    dictionary,
		feedbackStore: feedbackStore,
		miner:         miner,
		updater:       updater,
		requestLogger: requestLogger,
	}, nil
}

// SegmentRequest 分词请求
type SegmentRequest struct {
	Text string `json:"text"`
}

// SegmentResponse 分词响应
type SegmentResponse struct {
	Text   string      `json:"text"`
	Tokens []seg.Token `json:"tokens"`
	Words  []string    `json:"words"`
}

func handleSegment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 检查 mode 参数
	mode := r.URL.Query().Get("mode")
	var segmenter *seg.Segmenter
	if mode == "eval" {
		segmenter = engine.evalSegmenter
	} else {
		segmenter = engine.prodSegmenter
	}

	tokens := segmenter.Segment(req.Text)

	// 只过滤空白/不可见字符，保留标点符号
	filteredTokens := make([]seg.Token, 0, len(tokens))
	for _, t := range tokens {
		if strings.TrimSpace(t.Word) != "" {
			filteredTokens = append(filteredTokens, t)
		}
	}

	words := make([]string, len(filteredTokens))
	for i, t := range filteredTokens {
		words[i] = t.Word
	}

	resp := SegmentResponse{
		Text:   req.Text,
		Tokens: filteredTokens,
		Words:  words,
	}

	// 记录请求并检查是否应该触发自动挖掘
	shouldMine := engine.requestLogger.Add(req.Text)

	// 异步保存请求日志
	go engine.requestLogger.Save()

	if shouldMine {
		go func() {
			// 后台执行自动挖掘
			texts := engine.requestLogger.GetRecentTexts(500)
			result := engine.updater.AutoLearn(texts)
			if len(result.AddedWords) > 0 {
				engine.updater.SaveDictionary()
				log.Printf("Auto-mining: found %d new words: %v", len(result.AddedWords), result.AddedWords)
			}
			// 学习完成后清除请求记录
			engine.requestLogger.Clear()
			engine.requestLogger.Save()
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleSegmentSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 搜索总是使用 eval 模式，以便能搜索到新词
	tokens := engine.evalSegmenter.SegmentSearch(req.Text)

	resp := SegmentResponse{
		Text:   req.Text,
		Tokens: tokens,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// FeedbackRequest 反馈请求
type FeedbackRequest struct {
	Text         string   `json:"text"`
	OriginalSeg  []string `json:"original_seg"`
	CorrectedSeg []string `json:"corrected_seg"`
}

func handleFeedback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req FeedbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fb := engine.feedbackStore.Add(req.Text, req.OriginalSeg, req.CorrectedSeg)
		engine.feedbackStore.Save()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fb)

	case http.MethodGet:
		feedbacks := engine.feedbackStore.GetAll()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(feedbacks)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// WordRequest 词语请求
type WordRequest struct {
	Word string `json:"word"`
	Freq int    `json:"freq"`
}

func handleWords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		words := engine.dictionary.GetUserWords()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(words)

	case http.MethodPost:
		var req WordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Freq <= 0 {
			req.Freq = 100
		}
		engine.dictionary.AddWordToUser(req.Word, req.Freq)
		engine.dictionary.SaveDicts()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"word":    req.Word,
			"freq":    req.Freq,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleWordList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")
	dictType := r.URL.Query().Get("type") // "user", "base", "staging", "all"

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 {
		size = 50
	}

	type WordItem struct {
		Word string `json:"word"`
		Freq int    `json:"freq"`
		Type string `json:"type"`
	}

	var allWords []WordItem

	if dictType == "user" || dictType == "" || dictType == "all" {
		userWords := engine.dictionary.GetUserWords()
		for _, w := range userWords {
			allWords = append(allWords, WordItem{Word: w.Word, Freq: w.Freq, Type: "user"})
		}
	}

	if dictType == "base" || dictType == "all" {
		baseWords := engine.dictionary.GetBaseWords()
		for _, w := range baseWords {
			allWords = append(allWords, WordItem{Word: w.Word, Freq: w.Freq, Type: "base"})
		}
	}

	if dictType == "staging" || dictType == "all" {
		stagingWords := engine.dictionary.GetStagingWords()
		for _, w := range stagingWords {
			allWords = append(allWords, WordItem{Word: w.Word, Freq: w.Freq, Type: "staging"})
		}
	}

	// 按词频倒序排序
	sort.Slice(allWords, func(i, j int) bool {
		return allWords[i].Freq > allWords[j].Freq
	})

	total := len(allWords)
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}

	resp := map[string]interface{}{
		"total": total,
		"page":  page,
		"size":  size,
		"items": allWords[start:end],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleWordSearchAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		handleWordList(w, r)
		return
	}

	type WordItem struct {
		Word string `json:"word"`
		Freq int    `json:"freq"`
		Type string `json:"type"`
	}

	var results []WordItem

	userWords := engine.dictionary.GetUserWords()
	for _, w := range userWords {
		if strings.Contains(w.Word, q) {
			results = append(results, WordItem{Word: w.Word, Freq: w.Freq, Type: "user"})
		}
	}

	baseWords := engine.dictionary.GetBaseWords()
	for _, w := range baseWords {
		if strings.Contains(w.Word, q) {
			results = append(results, WordItem{Word: w.Word, Freq: w.Freq, Type: "base"})
		}
	}

	stagingWords := engine.dictionary.GetStagingWords()
	for _, w := range stagingWords {
		if strings.Contains(w.Word, q) {
			results = append(results, WordItem{Word: w.Word, Freq: w.Freq, Type: "staging"})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Freq > results[j].Freq
	})

	if len(results) > 100 {
		results = results[:100]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func handleWordOps(w http.ResponseWriter, r *http.Request) {
	// 提取词语
	word := strings.TrimPrefix(r.URL.Path, "/api/words/")
	if word == "" {
		http.Error(w, "Word is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Search in all dists (eval mode)
		freq, exists := engine.dictionary.Search(word, true)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"word":   word,
			"freq":   freq,
			"exists": exists,
		})

	case http.MethodDelete:
		success := engine.dictionary.RemoveWord(word)
		if success {
			engine.dictionary.SaveDicts()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": success,
			"word":    word,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LearnRequest 学习请求
type LearnRequest struct {
	Texts []string `json:"texts"`
	Type  string   `json:"type"` // "feedback" or "corpus"
}

func handleLearn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LearnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var result learn.UpdateResult

	switch req.Type {
	case "feedback":
		result = engine.updater.UpdateFromFeedback()
	case "corpus":
		if len(req.Texts) == 0 {
			http.Error(w, "Texts are required for corpus learning", http.StatusBadRequest)
			return
		}
		result = engine.updater.AutoLearn(req.Texts)
	default:
		// 默认处理反馈
		result = engine.updater.UpdateFromFeedback()
	}

	// 保存更新
	engine.updater.SaveDictionary()
	engine.updater.SaveFeedback()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleLearnFromRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取请求语料
	texts := engine.requestLogger.GetTexts()

	result := learn.UpdateResult{
		AddedWords:  make([]string, 0),
		UpdatedFreq: make(map[string]int),
		SourceTexts: len(texts),
	}

	if len(texts) > 0 {
		// 执行语料挖掘
		result = engine.updater.AutoLearn(texts)
		result.SourceTexts = len(texts)

		if len(result.AddedWords) > 0 {
			engine.updater.SaveDictionary()
		}

		// 学习完成后清除请求记录
		engine.requestLogger.Clear()
		engine.requestLogger.Save()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"dictionary": engine.dictionary.Stats(),
		"feedback":   engine.feedbackStore.Stats(),
		"requests":   engine.requestLogger.Stats(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleMergeDict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := engine.dictionary.MergeStagingToUser(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API Server Running. Please use the frontend application.")
}
