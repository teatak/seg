package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/teatak/seg/pkg/engine"
	"github.com/teatak/seg/pkg/learn"
	"github.com/teatak/seg/pkg/seg"
)

// Handler API处理器
type Handler struct {
	engine *engine.Engine
}

// NewHandler 创建新的API处理器
func NewHandler(e *engine.Engine) *Handler {
	return &Handler{engine: e}
}

// RegisterRoutes 注册路由到 ServeMux
// prefix: 路由前缀，例如 "/api"
func (h *Handler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	// 确保前缀不以 / 结尾，除非只有 /
	prefix = strings.TrimRight(prefix, "/")

	mux.HandleFunc(prefix+"/segment", h.handleSegment)
	mux.HandleFunc(prefix+"/segment/search", h.handleSegmentSearch)
	mux.HandleFunc(prefix+"/feedback", h.handleFeedback)
	mux.HandleFunc(prefix+"/words", h.handleWords)
	mux.HandleFunc(prefix+"/words/list", h.handleWordList)
	mux.HandleFunc(prefix+"/words/search", h.handleWordSearchAPI)
	mux.HandleFunc(prefix+"/words/", h.handleWordOps) // 注意这里会匹配 /api/words/{word}
	mux.HandleFunc(prefix+"/dict/merge", h.handleMergeDict)
	mux.HandleFunc(prefix+"/learn", h.handleLearn)
	mux.HandleFunc(prefix+"/learn/requests", h.handleLearnFromRequests)
	mux.HandleFunc(prefix+"/stats", h.handleStats)
}

// RegisterFrontend 注册前端静态文件服务 (SPA模式)
// distDir: 前端构建产物目录，例如 "./web/dist"
// apiPrefix: API 前缀，例如 "/api"，将注入到前端页面中
func (h *Handler) RegisterFrontend(mux *http.ServeMux, distDir string, apiPrefix string) {
	fs := http.FileServer(http.Dir(distDir))

	// 预加载并注入 index.html
	// 注意：这里假设 index.html 存在且较小，适合缓存在内存中
	// 如果需要热更，可以在 Handler 中每次读取
	var indexContent []byte
	// 简单的注入脚本
	injectScript := fmt.Sprintf("<script>window.API_PREFIX = \"%s\";</script>", apiPrefix)

	reloadIndex := func() {
		content, err := os.ReadFile(filepath.Join(distDir, "index.html"))
		if err == nil {
			// 插入到 <head> 之后，或者 <body> 之前
			// 简单起见，直接插入到 <head> 标签闭合前，如果没有则插入到 body 前
			sContent := string(content)
			if idx := strings.Index(sContent, "</head>"); idx != -1 {
				sContent = sContent[:idx] + injectScript + sContent[idx:]
			} else {
				sContent = injectScript + sContent
			}
			indexContent = []byte(sContent)
		} else {
			// 如果找不到文件，记录日志但不过分报错，可能尚未构建
			log.Printf("Warning: index.html not found in %s, frontend might not work correctly until built.", distDir)
			indexContent = nil
		}
	}

	// 初始加载
	reloadIndex()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// 如果请求的是 index.html 或者根路径，也返回注入后的内容
		if path == "" || path == "index.html" {
			if len(indexContent) == 0 {
				reloadIndex() // 尝试懒加载
				if len(indexContent) == 0 {
					http.Error(w, "Frontend not built (index.html missing)", http.StatusNotFound)
					return
				}
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexContent)
			return
		}

		// 尝试静态文件
		f, err := http.Dir(distDir).Open(path)
		if err == nil {
			defer f.Close()
			stat, _ := f.Stat()
			if !stat.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}

		// Fallback to index.html (SPA routing)
		if len(indexContent) == 0 {
			reloadIndex()
		}
		if len(indexContent) > 0 {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexContent)
		} else {
			http.NotFound(w, r)
		}
	})
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

func (h *Handler) handleSegment(w http.ResponseWriter, r *http.Request) {
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
		segmenter = h.engine.EvalSegmenter
	} else {
		segmenter = h.engine.ProdSegmenter
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
	shouldMine := h.engine.RequestLogger.Add(req.Text)

	// 异步保存请求日志
	go h.engine.RequestLogger.Save()

	if shouldMine {
		go func() {
			// 后台执行自动挖掘
			texts := h.engine.RequestLogger.GetRecentTexts(500)
			result := h.engine.Updater.AutoLearn(texts)
			if len(result.AddedWords) > 0 {
				h.engine.Updater.SaveDictionary()
				log.Printf("Auto-mining: found %d new words: %v", len(result.AddedWords), result.AddedWords)
			}
			// 学习完成后清除请求记录
			h.engine.RequestLogger.Clear()
			h.engine.RequestLogger.Save()
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleSegmentSearch(w http.ResponseWriter, r *http.Request) {
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
	tokens := h.engine.EvalSegmenter.SegmentSearch(req.Text)

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

func (h *Handler) handleFeedback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req FeedbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fb := h.engine.FeedbackStore.Add(req.Text, req.OriginalSeg, req.CorrectedSeg)
		h.engine.FeedbackStore.Save()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fb)

	case http.MethodGet:
		feedbacks := h.engine.FeedbackStore.GetAll()
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

func (h *Handler) handleWords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		words := h.engine.Dictionary.GetUserWords()
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
		h.engine.Dictionary.AddWordToUser(req.Word, req.Freq)
		h.engine.Dictionary.SaveDicts()

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

func (h *Handler) handleWordList(w http.ResponseWriter, r *http.Request) {
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
		userWords := h.engine.Dictionary.GetUserWords()
		for _, w := range userWords {
			allWords = append(allWords, WordItem{Word: w.Word, Freq: w.Freq, Type: "user"})
		}
	}

	if dictType == "base" || dictType == "all" {
		baseWords := h.engine.Dictionary.GetBaseWords()
		for _, w := range baseWords {
			allWords = append(allWords, WordItem{Word: w.Word, Freq: w.Freq, Type: "base"})
		}
	}

	if dictType == "staging" || dictType == "all" {
		stagingWords := h.engine.Dictionary.GetStagingWords()
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

func (h *Handler) handleWordSearchAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		h.handleWordList(w, r)
		return
	}

	type WordItem struct {
		Word string `json:"word"`
		Freq int    `json:"freq"`
		Type string `json:"type"`
	}

	var results []WordItem

	userWords := h.engine.Dictionary.GetUserWords()
	for _, w := range userWords {
		if strings.Contains(w.Word, q) {
			results = append(results, WordItem{Word: w.Word, Freq: w.Freq, Type: "user"})
		}
	}

	baseWords := h.engine.Dictionary.GetBaseWords()
	for _, w := range baseWords {
		if strings.Contains(w.Word, q) {
			results = append(results, WordItem{Word: w.Word, Freq: w.Freq, Type: "base"})
		}
	}

	stagingWords := h.engine.Dictionary.GetStagingWords()
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

func (h *Handler) handleWordOps(w http.ResponseWriter, r *http.Request) {
	// 提取词语
	// 路径格式: .../words/{word}
	parts := strings.Split(r.URL.Path, "/words/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	word := parts[len(parts)-1]

	if word == "" {
		http.Error(w, "Word is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Search in all dists (eval mode)
		freq, exists := h.engine.Dictionary.Search(word, true)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"word":   word,
			"freq":   freq,
			"exists": exists,
		})

	case http.MethodDelete:
		success := h.engine.Dictionary.RemoveWord(word)
		if success {
			h.engine.Dictionary.SaveDicts()
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

func (h *Handler) handleLearn(w http.ResponseWriter, r *http.Request) {
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
		result = h.engine.Updater.UpdateFromFeedback()
	case "corpus":
		if len(req.Texts) == 0 {
			http.Error(w, "Texts are required for corpus learning", http.StatusBadRequest)
			return
		}
		result = h.engine.Updater.AutoLearn(req.Texts)
	default:
		// 默认处理反馈
		result = h.engine.Updater.UpdateFromFeedback()
	}

	// 保存更新
	h.engine.Updater.SaveDictionary()
	h.engine.Updater.SaveFeedback()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) handleLearnFromRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取请求语料
	texts := h.engine.RequestLogger.GetTexts()

	result := learn.UpdateResult{
		AddedWords:  make([]string, 0),
		UpdatedFreq: make(map[string]int),
		SourceTexts: len(texts),
	}

	if len(texts) > 0 {
		// 执行语料挖掘
		result = h.engine.Updater.AutoLearn(texts)
		result.SourceTexts = len(texts)

		if len(result.AddedWords) > 0 {
			h.engine.Updater.SaveDictionary()
		}

		// 学习完成后清除请求记录
		h.engine.RequestLogger.Clear()
		h.engine.RequestLogger.Save()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"dictionary": h.engine.Dictionary.Stats(),
		"feedback":   h.engine.FeedbackStore.Stats(),
		"requests":   h.engine.RequestLogger.Stats(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) handleMergeDict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.engine.Dictionary.MergeStagingToUser(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
