package internal

import (
	"fmt"
	"regexp"
	"strings"
)

type AgentConfig struct {
	BaseURL     string
	ModelName   string
	MaxTokens   int
	Temperature float32
	ApiKey      string
}

type TranslationAgent struct {
	AgentConfig
}

func NewTranslationAgent(config AgentConfig) *TranslationAgent {
	return &TranslationAgent{config}
}

func (agent *TranslationAgent) Translate(sourceLang string, targetLang string, sourceText string, country string) string {
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
		for i := range translationChunks {
			translationChunks[i] = removeWrappingTags(translationChunks[i])
		}
		return strings.Trim(strings.Join(translationChunks, ""), "\n")
	}
}

// one chunk
func (agent *TranslationAgent) oneChunkInitialTranslation(sourceLang string, targetLang string, sourceText string) string {
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

func (agent *TranslationAgent) oneChunkReflectOnTranslation(sourceLang string, targetLang string, sourceText string, translation1 string, country string) string {
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

func (agent *TranslationAgent) oneChunkImproveTranslation(sourceLang string, targetLang string, sourceText string, translation1 string, reflection string) string {
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

func (agent *TranslationAgent) oneChunkTranslateText(sourceLang string, targetLang string, sourceText string, country string) string {
	translation1 := agent.oneChunkInitialTranslation(sourceLang, targetLang, sourceText)
	reflection := agent.oneChunkReflectOnTranslation(sourceLang, targetLang, sourceText, translation1, country)
	translation2 := agent.oneChunkImproveTranslation(sourceLang, targetLang, sourceText, translation1, reflection)
	return translation2
}

// multi chunk
func (agent *TranslationAgent) multiChunkInitialTranslation(sourceLang string, targetLang string, sourceTextChunks []string) []string {
	translationChunks := make([]string, len(sourceTextChunks))
	for i := range sourceTextChunks {
		taggedText := fmt.Sprintf("%s<TRANSLATE_THIS>%s</TRANSLATE_THIS>%s", strings.Join(sourceTextChunks[0:i], ""), sourceTextChunks[i], sourceTextChunks[i+1:])
		systemMessage, err := renderTemplate(multiChunkInitialTranslationSystemMessage, map[string]interface{}{
			"sourceLang": sourceLang,
			"targetLang": targetLang,
		})
		if err != nil {
			fmt.Printf("Error rendering initial translation system message: %v\n", err)
			return []string{}
		}
		translationPrompt, err := renderTemplate(multiChunkInitialTranslationPrompt, map[string]interface{}{
			"sourceLang":       sourceLang,
			"targetLang":       targetLang,
			"taggedText":       taggedText,
			"chunkToTranslate": sourceTextChunks[i],
		})
		if err != nil {
			fmt.Printf("Error rendering initial translation prompt: %v\n", err)
			return []string{}
		}
		translation, err := agent.getCompletion(translationPrompt, systemMessage)
		if err != nil {
			fmt.Printf("Error getting initial translation: %v\n", err)
			return []string{}
		}
		translationChunks[i] = translation
	}
	return translationChunks
}

func (agent *TranslationAgent) multiChunkReflectOnTranslation(sourceLang string, targetLang string, sourceTextChunks []string, translation1Chunks []string, country string) []string {
	reflectionChunks := make([]string, len(sourceTextChunks))
	for i := range sourceTextChunks {
		taggedText := fmt.Sprintf("%s<TRANSLATE_THIS>%s</TRANSLATE_THIS>%s", strings.Join(sourceTextChunks[0:i], ""), sourceTextChunks[i], sourceTextChunks[i+1:])
		systemMessage, err := renderTemplate(multiChunkReflectionSystemMessage, map[string]interface{}{
			"sourceLang": sourceLang,
			"targetLang": targetLang,
		})
		if err != nil {
			fmt.Printf("Error rendering reflection system message: %v\n", err)
			return []string{}
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
				return []string{}
			}
			reflection, err := agent.getCompletion(reflectionPrompt, systemMessage)
			if err != nil {
				fmt.Printf("Error getting reflection: %v\n", err)
				return []string{}
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
				return []string{}
			}
			reflection, err := agent.getCompletion(reflectionPrompt, systemMessage)
			if err != nil {
				fmt.Printf("Error getting reflection: %v\n", err)
				return []string{}
			}
			reflectionChunks[i] = reflection
		}
	}
	return reflectionChunks
}

func (agent *TranslationAgent) multiChunkImproveTranslation(sourceLang string, targetLang string, sourceTextChunks []string, translation1Chunks []string, reflectionChunks []string) []string {
	translation2Chunks := make([]string, len(sourceTextChunks))
	for i := range sourceTextChunks {
		taggedText := fmt.Sprintf("%s<TRANSLATE_THIS>%s</TRANSLATE_THIS>%s", strings.Join(sourceTextChunks[0:i], ""), sourceTextChunks[i], sourceTextChunks[i+1:])
		systemMessage, err := renderTemplate(multiChunkImproveTranslationSystemMessage, map[string]interface{}{
			"sourceLang": sourceLang,
			"targetLang": targetLang,
		})
		if err != nil {
			fmt.Printf("Error rendering improve translation system message: %v\n", err)
			return []string{}
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
			return []string{}
		}
		translation2, err := agent.getCompletion(improvementPrompt, systemMessage)
		if err != nil {
			fmt.Printf("Error getting improved translation: %v\n", err)
			return []string{}
		}
		translation2Chunks[i] = translation2
	}
	return translation2Chunks
}

func (agent *TranslationAgent) multiChunkTranslateText(sourceLang string, targetLang string, sourceTextChunks []string, country string) []string {
	translation1Chunks := agent.multiChunkInitialTranslation(sourceLang, targetLang, sourceTextChunks)
	reflectionChunks := agent.multiChunkReflectOnTranslation(sourceLang, targetLang, sourceTextChunks, translation1Chunks, country)
	translation2Chunks := agent.multiChunkImproveTranslation(sourceLang, targetLang, sourceTextChunks, translation1Chunks, reflectionChunks)
	return translation2Chunks
}

// removeWrappingTags Used to remove XML tags at the beginning and end
func removeWrappingTags(input string) string {
	// Matching forms like <TAG>... Start and end tags of </TAG>. Note that tag names are not captured to fit different tags
	re := regexp.MustCompile(`(?is)^<([a-zA-Z]+)[^>]*>(.*?)</\s*([a-zA-Z]+)\s*>$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 3 && strings.EqualFold(matches[1], matches[3]) {
		return matches[2]
	}
	return input
}
