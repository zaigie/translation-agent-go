package internal

import (
	"errors"
	"fmt"

	"github.com/pkoukk/tiktoken-go"
	ts "github.com/tmc/langchaingo/textsplitter"
)

var ErrNoSplitterNeeded = errors.New("input does not require splitting")

// createTextSplitter creates a text splitter using encoding or model name based on the input parameters.
func createTextSplitter(inputStr string, identifier string, maxTokens int, useModel bool) (ts.RecursiveCharacter, error) {
	var tokenEncoder *tiktoken.Tiktoken
	var err error

	if useModel {
		tokenEncoder, err = tiktoken.EncodingForModel(identifier)
	} else {
		tokenEncoder, err = tiktoken.GetEncoding(identifier)
	}
	if err != nil {
		return ts.RecursiveCharacter{}, fmt.Errorf("get encoding/model: %v", err)
	}

	numTokens := len(tokenEncoder.Encode(inputStr, nil, nil))
	if numTokens <= maxTokens {
		return ts.RecursiveCharacter{}, ErrNoSplitterNeeded
	}

	chunkSize := calculateChunkSize(numTokens, maxTokens)

	return ts.RecursiveCharacter{
		Separators:    []string{"。", "！", "？", "；", "……", "…", "\n\n", "\n", " ", ""},
		ChunkSize:     chunkSize,
		ChunkOverlap:  0,
		LenFunc:       func(s string) int { return len(tokenEncoder.Encode(s, nil, nil)) },
		KeepSeparator: true,
	}, nil
}

func calculateChunkSize(tokenCount int, tokenLimit int) int {
	// fmt.Printf("Number of tokens in text: %d, Token limit: %d\n", tokenCount, tokenLimit)
	if tokenCount <= tokenLimit {
		return tokenCount
	}
	numChunks := (tokenCount + tokenLimit - 1) / tokenLimit
	chunkSize := tokenCount / numChunks
	if remainingTokens := tokenCount % tokenLimit; remainingTokens > 0 {
		chunkSize += remainingTokens / numChunks
	}
	// fmt.Printf("Calculated chunk size: %d\n", chunkSize)
	return chunkSize
}
