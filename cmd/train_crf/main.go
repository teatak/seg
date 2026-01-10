package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/teatak/seg/optimizer"
)

func main() {
	inputPath := flag.String("input", "data/corpus.txt", "Path to the segmented corpus file")
	dictPath := flag.String("dict", "data/dict_base.txt", "Path to dictionary file (optional, each word becomes a training sentence)")
	outputPath := flag.String("output", "data/model.crf", "Path to save the model")
	iter := flag.Int("iter", 10, "Number of training iterations")
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Please provide input corpus using -input")
		os.Exit(1)
	}

	fmt.Printf("Training CRF model...\n")
	fmt.Printf("Input: %s\n", *inputPath)
	fmt.Printf("Dict Overlay: %s\n", *dictPath)
	fmt.Printf("Output: %s\n", *outputPath)
	fmt.Printf("Iterations: %d\n", *iter)

	err := optimizer.TrainCRF(*inputPath, *dictPath, *outputPath, *iter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Training failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully saved model to %s\n", *outputPath)
}
