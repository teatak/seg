package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/teatak/seg/crf"
	"github.com/teatak/seg/dictionary"
	"github.com/teatak/seg/optimizer"
	"github.com/teatak/seg/segmenter"
	"github.com/teatak/seg/util"
)

// Global segmenter with RWMutex for hot reloading
var (
	seg     *segmenter.Segmenter
	segLock sync.RWMutex
)

const (
	DataDir      = "data"
	LogFile      = "data/server_access.log"    // 沉淀用户输入
	NewWordsFile = "data/server_new_words.txt" // 挖掘出的新词
)

func main() {
	// 1. Initial Load
	if err := reloadEngine(); err != nil {
		log.Fatalf("Initial load failed: %v", err)
	}

	// 2. Setup Log file
	logF, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logF.Close()

	// 3. Handlers
	http.Handle("/", http.FileServer(http.Dir("./static"))) // 静态前端
	http.HandleFunc("/segment", func(w http.ResponseWriter, r *http.Request) {
		handleSegment(w, r, logF)
	})
	http.HandleFunc("/feedback", handleFeedback)         // 人工教词
	http.HandleFunc("/trigger-discovery", handleTrigger) // 触发自动挖掘 & 训练

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// reloadEngine reloads dict and model from disk safely
func reloadEngine() error {
	log.Println("Reloading engine...")
	dict := dictionary.NewDictionary()

	// Load hierarchical dictionaries in order: Core -> Base -> User
	// (Last one loaded wins frequency and existence)
	paths := []struct {
		name string
		path string
	}{
		{"Core", "data/dict_core.txt"},
		{"Base", "data/dict_base.txt"},
		{"User", "data/dict_user.txt"},
	}

	for _, d := range paths {
		if util.FileExists(d.path) {
			if err := dict.Load(d.path); err != nil {
				log.Printf("Error loading %s dictionary: %v", d.name, err)
			} else {
				log.Printf("Loaded %s dictionary.", d.name)
			}
		} else {
			log.Printf("Note: %s dictionary not found.", d.name)
		}
	}

	newSeg := segmenter.NewSegmenter(dict)
	model := crf.NewModel()
	if util.FileExists("data/model.crf") {
		if err := model.Load("data/model.crf"); err == nil {
			newSeg.CRFModel = model
		} else {
			log.Printf("Error loading CRF model: %v", err)
		}
	} else {
		log.Println("Warning: No CRF model found, running in pure DAG mode.")
	}

	segLock.Lock()
	seg = newSeg
	segLock.Unlock()
	log.Println("Engine reloaded successfully.")
	return nil
}

// Request/Response types
type SegRequest struct {
	Text      string `json:"text"`
	Function  string `json:"function"`  // standard, search
	Algorithm string `json:"algorithm"` // hybrid, crf, dag
}

type SegResponse struct {
	Tokens []string `json:"tokens"`
}

func handleSegment(w http.ResponseWriter, r *http.Request, logFile io.Writer) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req SegRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// 1. Log input for future discovery (Async write)
	go func(text string) {
		// Simple text cleaning could happen here
		if len(text) > 2 {
			logFile.Write([]byte(text + "\n"))
		}
	}(req.Text)

	// 2. Process
	segLock.RLock()
	s := seg
	segLock.RUnlock()

	// Resolve Algorithm Mode
	var segMode segmenter.Mode
	switch req.Algorithm {
	case "crf":
		segMode = segmenter.ModeCRF
	case "dag":
		segMode = segmenter.ModeDAG
	default:
		segMode = segmenter.ModeHybrid
	}

	var tokens []string
	if req.Function == "search" {
		tokens = s.CutSearch(req.Text, segMode)
	} else {
		tokens = s.Cut(req.Text, segMode)
	}

	json.NewEncoder(w).Encode(SegResponse{Tokens: tokens})
}

func handleFeedback(w http.ResponseWriter, r *http.Request) {
	// User explicitly tells us a new word (or words if split by space)
	rawInput := r.URL.Query().Get("word")
	if rawInput == "" {
		http.Error(w, "word param required", 400)
		return
	}

	// Handle multiple words (e.g. "Hello World" -> "Hello", "World")
	words := strings.Fields(rawInput)
	if len(words) == 0 {
		http.Error(w, "empty input", 400)
		return
	}

	// Append to new words file
	f, err := os.OpenFile(NewWordsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()

	// Write the full raw user input as a single line to preserve context (e.g. "A B" implies split).
	if _, err := f.WriteString(rawInput + "\n"); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Trigger optimize immediately? Or let user trigger separately.
	// Let's trigger immediately for "instant feedback" feel.
	go runOptimization()

	action := "added"
	if len(words) > 1 {
		action = "split and added"
	}
	fmt.Fprintf(w, "Words %v. Optimization started in background.", action)
}

func handleTrigger(w http.ResponseWriter, r *http.Request) {
	// 1. Run Discovery on server_access.log
	log.Println("Running unsupervised discovery on access logs...")

	// Create a temp file for discovered words
	tmpDiscovered := "data/temp_discovered.txt"

	// Call internal optimizer discovery
	// threshold=3, ngram=4
	if err := optimizer.Discover(LogFile, tmpDiscovered, 3, 4); err != nil {
		log.Printf("Discovery failed: %v", err)
		http.Error(w, fmt.Sprintf("Discovery failed: %v", err), 500)
		return
	}

	// 2. Append discovered words to NewWordsFile
	content, err := os.ReadFile(tmpDiscovered)
	if err == nil && len(content) > 0 {
		f, err := os.OpenFile(NewWordsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			f.Write(content)
			f.Close()
			log.Printf("Appended discovered words to %s", NewWordsFile)
		}
	}

	// Clean up temp file
	os.Remove(tmpDiscovered)

	// Truncate access log after processing so we don't re-process old data
	if err := os.Truncate(LogFile, 0); err != nil {
		log.Printf("Warning: Failed to truncate log file: %v", err)
	} else {
		log.Printf("Truncated discovery source file %s", LogFile)
	}

	// 3. Run Optimization
	go runOptimization()
	fmt.Fprintln(w, "Discovery completed. Optimization pipeline started in background.")
}

func runOptimization() {
	log.Println("Starting optimization pipeline...")

	// Call internal optimizer pipeline directly
	if err := optimizer.Run(NewWordsFile); err != nil {
		log.Printf("Optimization failed: %v", err)
		return
	}
	log.Printf("Optimization finished.")

	// Clear the new words file so we don't re-add them next time?
	// Actually optimizer.Run merges them into main dict.
	// We MUST truncate it otherwise optimizer.Run will re-add them with higher freq again and again.
	os.Truncate(NewWordsFile, 0)

	// Reload Engine
	reloadEngine()
}
