package main

import (
	"fmt"

	ta "github.com/zaigie/translation-agent-go"
)

func main() {
	text := "真空吸附磁吸手机车载支架，轻轻一放，就能牢牢吸住的手机支架，不挑车型和手机，想吸哪里吸哪里。双向吸附设计，一端是电动真空吸附，一端是强磁吸附，两种模式配合，适合多种场景使用。采用高品质强力磁铁，确保手机在各种颠簸路面和日常使用中的稳定性，不易脱落。能吸力感知，吸力变弱时自动重启吸附，时刻保持最佳吸附状态，无需担心设备掉落。广角度支架轻松调整最合适的角度，不管什么角度都能获得最佳视角体验。磁吸端自带加厚硅胶软垫，有效保护免受刮擦和磨损。同时硅胶材质柔软，有效减少震动，提升使用稳定性"
	agent := ta.NewTranslateAgent(ta.AgentConfig{
		BaseURL:     "https://apivip.aiproxy.io/v1",
		ModelName:   "gpt-4o-mini",
		MaxTokens:   1000,
		Temperature: 0.3,
		ApiKey:      "sk-xxxxxx",
	})

	translatedText := agent.Translate("Chinese", "English", text, "America")
	fmt.Printf("Translated text: %s\n", translatedText)
}
