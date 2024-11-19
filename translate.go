package internal

import (
	"fmt"
	"strings"
)

type AgentConfig struct {
	BaseURL     string
	ModelName   string
	MaxTokens   int
	Temperature float32
	ApiKey      string
}

type TranslateAgent struct {
	AgentConfig
}

func NewTranslateAgent(config AgentConfig) *TranslateAgent {
	return &TranslateAgent{config}
}

func (agent *TranslateAgent) Translate(sourceLang string, targetLang string, sourceText string, country string) string {
	// Check if the source text is a single chunk or multiple chunks
	// If the source text is a single chunk, call the OneChunkTranslateText function
	// If the source text is multiple chunks, call the MultiChunkTranslateText function
	textSplitter, err := createTextSplitter(sourceText, agent.ModelName, agent.MaxTokens, true)
	if err != nil {
		if err != ErrNoSplitterNeeded {
			fmt.Printf("Error creating text splitter: %v\n", err)
			return ""
		}
	}
	if err == ErrNoSplitterNeeded {
		return agent.oneChunkTranslateText(sourceLang, targetLang, sourceText, country)
	} else {
		textChunks, err := textSplitter.SplitText(sourceText)
		if err != nil {
			fmt.Printf("Error splitting text: %v\n", err)
			return ""
		}
		translationChunks := agent.multiChunkTranslateText(sourceLang, targetLang, textChunks, country)
		return strings.Join(translationChunks, "")
	}
}

// one chunk
func (agent *TranslateAgent) oneChunkInitialTranslation(sourceLang string, targetLang string, sourceText string) string {
	systemMessage, err := renderTemplate(oneChunkInitialTranslationSystemMessage, map[string]interface{}{
		"sourceLang": sourceLang,
		"targetLang": targetLang,
	})
	if err != nil {
		fmt.Printf("Error rendering initial translation system message: %v\n", err)
		return ""
	}
	translationPrompt, err := renderTemplate(oneChunkInitialTranslationPrompt, map[string]interface{}{
		"sourceLang": sourceLang,
		"sourceText": sourceText,
		"targetLang": targetLang,
	})
	if err != nil {
		fmt.Printf("Error rendering initial translation prompt: %v\n", err)
		return ""
	}
	translation, err := agent.getCompletion(translationPrompt, systemMessage)
	if err != nil {
		fmt.Printf("Error getting initial translation: %v\n", err)
		return ""
	}
	return translation
}

func (agent *TranslateAgent) oneChunkReflectOnTranslation(sourceLang string, targetLang string, sourceText string, translation1 string, country string) string {
	systemMessage, err := renderTemplate(oneChunkReflectionSystemMessage, map[string]interface{}{
		"sourceLang": sourceLang,
		"targetLang": targetLang,
	})
	if err != nil {
		fmt.Printf("Error rendering reflection system message: %v\n", err)
		return ""
	}
	var reflectionPrompt string = ""
	if country == "" {
		reflectionPrompt, err = renderTemplate(oneChunkReflectionPrompt, map[string]interface{}{
			"sourceLang":   sourceLang,
			"sourceText":   sourceText,
			"translation1": translation1,
			"targetLang":   targetLang,
		})
		if err != nil {
			fmt.Printf("Error rendering reflection prompt: %v\n", err)
			return ""
		}
	} else {
		reflectionPrompt, err = renderTemplate(oneChunkReflectionCountryPrompt, map[string]interface{}{
			"sourceLang":   sourceLang,
			"sourceText":   sourceText,
			"translation1": translation1,
			"targetLang":   targetLang,
			"country":      country,
		})
		if err != nil {
			fmt.Printf("Error rendering reflection prompt: %v\n", err)
			return ""
		}
	}
	reflection, err := agent.getCompletion(reflectionPrompt, systemMessage)
	if err != nil {
		fmt.Printf("Error getting reflection: %v\n", err)
		return ""
	}
	return reflection
}

func (agent *TranslateAgent) oneChunkImproveTranslation(sourceLang string, targetLang string, sourceText string, translation1 string, reflection string) string {
	systemMessage, err := renderTemplate(oneChunkImproveTranslationSystemMessage, map[string]interface{}{
		"sourceLang": sourceLang,
		"targetLang": targetLang,
	})
	if err != nil {
		fmt.Printf("Error rendering improve translation system message: %v\n", err)
		return ""
	}
	improvementPrompt, err := renderTemplate(oneChunkImproveTranslationPrompt, map[string]interface{}{
		"sourceLang":   sourceLang,
		"sourceText":   sourceText,
		"translation1": translation1,
		"reflection":   reflection,
		"targetLang":   targetLang,
	})
	if err != nil {
		fmt.Printf("Error rendering improve translation prompt: %v\n", err)
		return ""
	}
	translation2, err := agent.getCompletion(improvementPrompt, systemMessage)
	if err != nil {
		fmt.Printf("Error getting improved translation: %v\n", err)
		return ""
	}
	return translation2
}

func (agent *TranslateAgent) oneChunkTranslateText(sourceLang string, targetLang string, sourceText string, country string) string {
	translation1 := agent.oneChunkInitialTranslation(sourceLang, targetLang, sourceText)
	reflection := agent.oneChunkReflectOnTranslation(sourceLang, targetLang, sourceText, translation1, country)
	translation2 := agent.oneChunkImproveTranslation(sourceLang, targetLang, sourceText, translation1, reflection)
	return translation2
}

// multi chunk
func (agent *TranslateAgent) multiChunkInitialTranslation(sourceLang string, targetLang string, sourceTextChunks []string) []string {
	translationChunks := make([]string, len(sourceTextChunks))
	for i := range sourceTextChunks {
		taggedText := fmt.Sprintf("%s<TRANSLATE_THIS>%s</TRANSLATE_THIS>%s", strings.Join(sourceTextChunks[0:i], ""), sourceTextChunks[i], sourceTextChunks[i+1:])
		systemMessage, err := renderTemplate(multiChunkInitialTranslationSystemMessage, map[string]interface{}{
			"sourceLang": sourceLang,
			"targetLang": targetLang,
		})
		if err != nil {
			fmt.Printf("Error rendering initial translation system message: %v\n", err)
			return nil
		}
		translationPrompt, err := renderTemplate(multiChunkInitialTranslationPrompt, map[string]interface{}{
			"sourceLang":       sourceLang,
			"targetLang":       targetLang,
			"taggedText":       taggedText,
			"chunkToTranslate": sourceTextChunks[i],
		})
		if err != nil {
			fmt.Printf("Error rendering initial translation prompt: %v\n", err)
			return nil
		}
		translation, err := agent.getCompletion(translationPrompt, systemMessage)
		if err != nil {
			fmt.Printf("Error getting initial translation: %v\n", err)
			return nil
		}
		translationChunks[i] = translation
	}
	return translationChunks
}

func (agent *TranslateAgent) multiChunkReflectOnTranslation(sourceLang string, targetLang string, sourceTextChunks []string, translation1Chunks []string, country string) []string {
	reflectionChunks := make([]string, len(sourceTextChunks))
	for i := range sourceTextChunks {
		taggedText := fmt.Sprintf("%s<TRANSLATE_THIS>%s</TRANSLATE_THIS>%s", strings.Join(sourceTextChunks[0:i], ""), sourceTextChunks[i], sourceTextChunks[i+1:])
		systemMessage, err := renderTemplate(multiChunkReflectionSystemMessage, map[string]interface{}{
			"sourceLang": sourceLang,
			"targetLang": targetLang,
		})
		if err != nil {
			fmt.Printf("Error rendering reflection system message: %v\n", err)
			return nil
		}
		if country == "" {
			reflectionPrompt, err := renderTemplate(multiChunkReflectionPrompt, map[string]interface{}{
				"sourceLang":        sourceLang,
				"targetLang":        targetLang,
				"taggedText":        taggedText,
				"chunkToTranslate":  sourceTextChunks[i],
				"translation1Chunk": translation1Chunks[i],
			})
			if err != nil {
				fmt.Printf("Error rendering reflection prompt: %v\n", err)
				return nil
			}
			reflection, err := agent.getCompletion(reflectionPrompt, systemMessage)
			if err != nil {
				fmt.Printf("Error getting reflection: %v\n", err)
				return nil
			}
			reflectionChunks[i] = reflection
		} else {
			reflectionPrompt, err := renderTemplate(multiChunkReflectionCountryPrompt, map[string]interface{}{
				"sourceLang":        sourceLang,
				"targetLang":        targetLang,
				"taggedText":        taggedText,
				"chunkToTranslate":  sourceTextChunks[i],
				"translation1Chunk": translation1Chunks[i],
				"country":           country,
			})
			if err != nil {
				fmt.Printf("Error rendering reflection prompt: %v\n", err)
				return nil
			}
			reflection, err := agent.getCompletion(reflectionPrompt, systemMessage)
			if err != nil {
				fmt.Printf("Error getting reflection: %v\n", err)
				return nil
			}
			reflectionChunks[i] = reflection
		}
	}
	return reflectionChunks
}

func (agent *TranslateAgent) multiChunkImproveTranslation(sourceLang string, targetLang string, sourceTextChunks []string, translation1Chunks []string, reflectionChunks []string) []string {
	translation2Chunks := make([]string, len(sourceTextChunks))
	for i := range sourceTextChunks {
		taggedText := fmt.Sprintf("%s<TRANSLATE_THIS>%s</TRANSLATE_THIS>%s", strings.Join(sourceTextChunks[0:i], ""), sourceTextChunks[i], sourceTextChunks[i+1:])
		systemMessage, err := renderTemplate(multiChunkImproveTranslationSystemMessage, map[string]interface{}{
			"sourceLang": sourceLang,
			"targetLang": targetLang,
		})
		if err != nil {
			fmt.Printf("Error rendering improve translation system message: %v\n", err)
			return nil
		}
		improvementPrompt, err := renderTemplate(multiChunkImproveTranslationPrompt, map[string]interface{}{
			"sourceLang":        sourceLang,
			"targetLang":        targetLang,
			"taggedText":        taggedText,
			"chunkToTranslate":  sourceTextChunks[i],
			"translation1Chunk": translation1Chunks[i],
			"reflectionChunk":   reflectionChunks[i],
		})
		if err != nil {
			fmt.Printf("Error rendering improve translation prompt: %v\n", err)
			return nil
		}
		translation2, err := agent.getCompletion(improvementPrompt, systemMessage)
		if err != nil {
			fmt.Printf("Error getting improved translation: %v\n", err)
			return nil
		}
		translation2Chunks[i] = translation2
	}
	return translation2Chunks
}

func (agent *TranslateAgent) multiChunkTranslateText(sourceLang string, targetLang string, sourceTextChunks []string, country string) []string {
	translation1Chunks := agent.multiChunkInitialTranslation(sourceLang, targetLang, sourceTextChunks)
	reflectionChunks := agent.multiChunkReflectOnTranslation(sourceLang, targetLang, sourceTextChunks, translation1Chunks, country)
	translation2Chunks := agent.multiChunkImproveTranslation(sourceLang, targetLang, sourceTextChunks, translation1Chunks, reflectionChunks)
	return translation2Chunks
}
