# seg

[中文](README_ZH.md)

A lightweight and high-performance Chinese segmentation library in Go.

## Features

- **Dictionary-based segmentation**: Precise segmentation using custom dictionaries.
- **Hybrid Mode**: Combines the precision of dictionary lookups with the adaptability of CRF models to handle both known and unknown words effectively.
- **CRF Model Support**: Integrated Linear-chain Conditional Random Field (CRF) for intelligent Out-of-Vocabulary (OOV) word recognition.
- **Efficient Algorithms**: Uses Directed Acyclic Graph (DAG) for candidate generation and Dynamic Programming (DP) to find the maximum probability path.
- **Alphanumeric Protection**: Automatically preserves alphanumeric sequences (e.g., "7天", "iPhone15", "PKU") without unnecessary fragmentation.

## Installation

```bash
go get github.com/teatak/seg
```

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/teatak/seg/crf"
	"github.com/teatak/seg/dictionary"
	"github.com/teatak/seg/segmenter"
)

func main() {
	// 1. Load Dictionary
	dict := dictionary.NewDictionary()
	dict.Load("data/dictionary.txt")

	// 2. Initialize Segmenter
	seg := segmenter.NewSegmenter(dict)

	// 3. Load CRF Model (Required for ModeHybrid)
	model := crf.NewModel()
	if err := model.Load("data/crf_model.txt"); err == nil {
		seg.CRFModel = model
	}

	text := "我想去看看人工智能的发展"

	// 4. Hybrid Segmentation (Recommended)
	// Combines dictionary lookup with CRF discovery
	parts := seg.Cut(text, segmenter.ModeHybrid)
	fmt.Println("Hybrid:", parts) 

	// 5. Search Engine Mode (Fine-grained)
	// Useful for indexing
	searchParts := seg.CutSearch(text, segmenter.ModeHybrid)
	fmt.Println("Search:", searchParts)
}
```

## Segmentation Modes

- **`segmenter.ModeHybrid` (Recommended)**: Best of both worlds. Locks words found in the dictionary and uses CRF to predict the remaining unknown segments.
- **`segmenter.ModeDAG`**: Strict dictionary-based segmentation. Fast but cannot recognize new words.
- **`segmenter.ModeCRF`**: Pure model-based prediction. Good for research or when no dictionary is available, but may be unstable on known words.

## CLI Tool

Test the segmentation directly from your terminal:

```bash
# Standard Hybrid Mode (Default)
go run cmd/seg/main.go "丽怡酒店的价格是否包含早餐"

# Search Engine Mode
go run cmd/seg/main.go -func search "人工智能"

# Pure CRF Mode
go run cmd/seg/main.go -mode crf "丽怡酒店的价格"
```

## License

MIT License