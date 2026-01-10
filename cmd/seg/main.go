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
	"github.com/teatak/seg/util"
)

func main() {
	function := flag.String("func", "cut", "Segmentation function: cut (standard) or search (for search engine)")
	mode := flag.String("mode", "hybrid", "Algorithm mode: hybrid (recommended), dag, or crf")
	basePath := flag.String("base", "data/dict_base.txt", "Path to base dictionary")
	corePath := flag.String("core", "data/dict_core.txt", "Path to core dictionary")
	userPath := flag.String("user", "data/dict_user.txt", "Path to user dictionary")
	modelPath := flag.String("model", "data/model.crf", "Path to CRF model file")
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

	// 2. Load Resources (Dict / Model)
	dict := dictionary.NewDictionary()

	// Load hierarchical dictionaries in order: Core -> Base -> User
	// (Last one loaded wins frequency and existence)
	if util.FileExists(*corePath) {
		dict.Load(*corePath)
	}
	if util.FileExists(*basePath) {
		dict.Load(*basePath)
	}
	if util.FileExists(*userPath) {
		dict.Load(*userPath)
	}

	if dict.Total == 0 && *mode != "crf" {
		fmt.Fprintf(os.Stderr, "Warning: No dictionary loaded. Standard mode may be inaccurate.\n")
	}

	seg := segmenter.NewSegmenter(dict)

	// Load CRF Model
	// Required for: crf
	// Recommended for: hybrid
	if *mode == "crf" || *mode == "hybrid" || util.FileExists(*modelPath) {
		if util.FileExists(*modelPath) {
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
