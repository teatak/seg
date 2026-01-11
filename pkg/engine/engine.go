package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/teatak/seg/pkg/dict"
	"github.com/teatak/seg/pkg/learn"
	"github.com/teatak/seg/pkg/seg"
)

// Config 配置
type Config struct {
	DataDir     string
	BaseDict    string
	UserDict    string
	StagingDict string
	FeedbackDB  string
	RequestLog  string
}

// Engine 分词引擎核心
type Engine struct {
	ProdSegmenter *seg.Segmenter // Production segmenter
	EvalSegmenter *seg.Segmenter // Evaluation segmenter
	Dictionary    *dict.Dictionary
	FeedbackStore *learn.FeedbackStore
	Miner         *learn.Miner
	Updater       *learn.Updater
	RequestLogger *learn.RequestLogger
}

// NewEngine 创建引擎
func NewEngine(cfg Config) (*Engine, error) {
	// 默认路径处理
	if cfg.BaseDict == "" {
		cfg.BaseDict = filepath.Join(cfg.DataDir, "dict/base.txt")
	}
	if cfg.UserDict == "" {
		cfg.UserDict = filepath.Join(cfg.DataDir, "dict/user.txt")
	}
	if cfg.StagingDict == "" {
		cfg.StagingDict = filepath.Join(cfg.DataDir, "dict/staging.txt")
	}
	if cfg.FeedbackDB == "" {
		cfg.FeedbackDB = filepath.Join(cfg.DataDir, "feedback.json")
	}
	if cfg.RequestLog == "" {
		cfg.RequestLog = filepath.Join(cfg.DataDir, "requests.json")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(cfg.UserDict), 0755); err != nil {
		return nil, fmt.Errorf("failed to create dict dir: %w", err)
	}

	// 创建词典 (Init with staging dict)
	dictionary := dict.NewDictionary(cfg.BaseDict, cfg.UserDict, cfg.StagingDict)
	if err := dictionary.Load(); err != nil {
		return nil, fmt.Errorf("failed to load dictionary: %w", err)
	}

	// 创建 HMM 模型
	var hmm *seg.HMM
	hmmFile := filepath.Join(cfg.DataDir, "model/hmm.json")
	if _, err := os.Stat(hmmFile); err == nil {
		hmm = seg.NewHMM()
		if err := hmm.LoadFromFile(hmmFile); err != nil {
			log.Printf("Warning: Failed to load HMM model from %s: %v. Using default.", hmmFile, err)
			hmm = seg.NewDefaultHMM()
		} else {
			log.Printf("Loaded custom HMM model from %s", hmmFile)
		}
	} else {
		hmm = seg.NewDefaultHMM()
	}

	// 创建分词器 (Dual instances)
	prodSegmenter := seg.NewSegmenter(dictionary, hmm, false)
	evalSegmenter := seg.NewSegmenter(dictionary, hmm, true)

	// 创建反馈存储
	feedbackStore := learn.NewFeedbackStore(cfg.FeedbackDB)
	feedbackStore.Load()

	// 创建新词发现器
	miner := learn.NewMiner()

	// 创建更新器
	updater := learn.NewUpdater(dictionary, feedbackStore, miner, evalSegmenter)

	// 创建请求记录器
	requestLogger := learn.NewRequestLogger(cfg.RequestLog)
	requestLogger.Load()

	return &Engine{
		ProdSegmenter: prodSegmenter,
		EvalSegmenter: evalSegmenter,
		Dictionary:    dictionary,
		FeedbackStore: feedbackStore,
		Miner:         miner,
		Updater:       updater,
		RequestLogger: requestLogger,
	}, nil
}
