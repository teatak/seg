package main

import (
	"fmt"
	"log"

	"github.com/teatak/seg/crf"
	"github.com/teatak/seg/dictionary"
	"github.com/teatak/seg/segmenter"
)

func main() {
	// 1. Create a segmenter (Dictionary is optional for CutCRF if you don't use DAG)
	// But Segmenter struct currently requires Dict potentially for other methods.
	// You can pass an empty dictionary if you only want to use CutCRF.
	dict := dictionary.NewDictionary()
	seg := segmenter.NewSegmenter(dict)

	// 2. Load CRF Model
	model := crf.NewModel()
	err := model.Load("data/crf_model.txt")
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}
	seg.CRFModel = model

	// 3. Segment
	cases := []string{
		"我们是程序员",
		"南京市长江大桥",
		"丽怡酒店",
		"茂名高铁站",
		"7天优品酒店",
		"武汉汉口火车站",
	}

	for _, text := range cases {
		parts := seg.Cut(text, segmenter.ModeCRF)
		fmt.Printf("Input: %s\nOutput: %v\n", text, parts)
	}
}
