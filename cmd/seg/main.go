package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/teatak/seg/crf"
	"github.com/teatak/seg/dictionary"
	"github.com/teatak/seg/segmenter"
)

func main() {
	function := flag.String("func", "cut", "Segmentation function: cut (standard) or search (for search engine)")
	mode := flag.String("mode", "hybrid", "Algorithm mode: hybrid (recommended), dag, or crf")
	dictPath := flag.String("dict", "data/dictionary.txt", "Path to dictionary file")
	modelPath := flag.String("model", "data/crf_model.txt", "Path to CRF model file")
	flag.Parse()

	// 1. Resolve Mode Constant
	var segMode segmenter.Mode
	switch *mode {
	case "hybrid":
		segMode = segmenter.ModeHybrid
	case "crf":
		segMode = segmenter.ModeCRF
	case "dag":
		segMode = segmenter.ModeDAG
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode '%s'. Using 'hybrid'.\n", *mode)
		segMode = segmenter.ModeHybrid
		*mode = "hybrid" // Normalize for loading logic
	}

	// 2. Load Resources (Dict / Model) based on Algorithm Mode

	// Load Dictionary
	// Required for: hybrid, dag
	// Optional for: crf
	var dict *dictionary.Dictionary
	if *mode != "crf" {
		if !fileExists(*dictPath) {
			fmt.Fprintf(os.Stderr, "Error: Dictionary file not found at %s. Required for mode '%s'.\n", *dictPath, *mode)
			os.Exit(1)
		}
		dict = dictionary.NewDictionary()
		if err := dict.Load(*dictPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading dictionary: %v\n", err)
			os.Exit(1)
		}
	} else {
		// CRF mode
		if fileExists(*dictPath) {
			dict = dictionary.NewDictionary()
			_ = dict.Load(*dictPath)
		} else {
			dict = dictionary.NewDictionary()
		}
	}

	seg := segmenter.NewSegmenter(dict)

	// Load CRF Model
	// Required for: crf
	// Recommended for: hybrid
	if *mode == "crf" || *mode == "hybrid" || fileExists(*modelPath) {
		if fileExists(*modelPath) {
			crfModel := crf.NewModel()
			if err := crfModel.Load(*modelPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading CRF model: %v\n", err)
				if *mode == "crf" {
					os.Exit(1)
				}
			} else {
				seg.CRFModel = crfModel
			}
		} else {
			if *mode == "crf" {
				fmt.Fprintf(os.Stderr, "Error: CRF model file not found at %s. Required for mode 'crf'.\n", *modelPath)
				os.Exit(1)
			}
			if *mode == "hybrid" {
				fmt.Fprintf(os.Stderr, "Warning: CRF model not found at %s. Downgrading 'hybrid' to DAG-only.\n", *modelPath)
			}
		}
	}

	// Helper to process text
	process := func(text string) []string {
		// Dispatch based on Function
		if *function == "search" {
			return seg.CutSearch(text, segMode)
		}
		// Default to cut
		return seg.Cut(text, segMode)
	}

	// If args provided (non-flag args), segment them
	args := flag.Args()
	if len(args) > 0 {
		text := strings.Join(args, " ")
		result := process(text)
		fmt.Println(strings.Join(result, " / "))
		return
	}

	// Otherwise interactive mode
	fmt.Println("Enter text to segment (Ctrl+D to exit):")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		result := process(text)
		fmt.Println(strings.Join(result, " / "))
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
