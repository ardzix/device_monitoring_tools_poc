package analysis

import (
	"math/rand"
	"strings"
	"time"
)

type AnalysisResult struct {
	Keywords    []string
	Confidence  float64
	IsFlagged   bool
	Description string
}

type FakeAnalyzer struct {
	// Predefined keyword sets for different contexts
	sensitiveKeywords []string
	actionKeywords    []string
	contextKeywords   map[string][]string
}

func NewFakeAnalyzer() *FakeAnalyzer {
	return &FakeAnalyzer{
		sensitiveKeywords: []string{
			"confidential",
			"password",
			"secret",
			"private",
			"classified",
			"restricted",
			"internal",
			"sensitive",
		},
		actionKeywords: []string{
			"delete",
			"modify",
			"create",
			"execute",
			"install",
			"download",
			"upload",
			"share",
		},
		contextKeywords: map[string][]string{
			"browser": {
				"browsing",
				"web",
				"internet",
				"online",
				"website",
				"search",
			},
			"terminal": {
				"command",
				"shell",
				"script",
				"sudo",
				"root",
				"system",
			},
			"editor": {
				"editing",
				"code",
				"development",
				"programming",
				"writing",
				"document",
			},
			"chat": {
				"message",
				"communication",
				"chat",
				"conversation",
				"social",
				"meeting",
			},
		},
	}
}

func (a *FakeAnalyzer) generateContextKeywords(windowTitle string) []string {
	title := strings.ToLower(windowTitle)
	var keywords []string

	// Add context-based keywords
	if strings.Contains(title, "chrome") || strings.Contains(title, "firefox") || strings.Contains(title, "edge") {
		keywords = append(keywords, a.contextKeywords["browser"]...)
	}
	if strings.Contains(title, "terminal") || strings.Contains(title, "cmd") || strings.Contains(title, "powershell") {
		keywords = append(keywords, a.contextKeywords["terminal"]...)
	}
	if strings.Contains(title, "code") || strings.Contains(title, "vim") || strings.Contains(title, "notepad") {
		keywords = append(keywords, a.contextKeywords["editor"]...)
	}
	if strings.Contains(title, "teams") || strings.Contains(title, "slack") || strings.Contains(title, "discord") {
		keywords = append(keywords, a.contextKeywords["chat"]...)
	}

	return keywords
}

func (a *FakeAnalyzer) AnalyzeScreenshot(imagePath string) AnalysisResult {
	// Extract window title from image path for context
	parts := strings.Split(imagePath, "/")
	filename := parts[len(parts)-1]
	windowTitle := strings.ToLower(filename)

	// Simulate random analysis with context-aware keywords
	rand.Seed(time.Now().UnixNano())

	// Generate base keywords from context
	contextKeywords := a.generateContextKeywords(windowTitle)

	// 30% chance of finding something sensitive
	if rand.Float64() < 0.3 {
		// Select 1-2 sensitive keywords
		numSensitive := rand.Intn(2) + 1
		for i := 0; i < numSensitive; i++ {
			keyword := a.sensitiveKeywords[rand.Intn(len(a.sensitiveKeywords))]
			contextKeywords = append(contextKeywords, keyword)
		}

		// Add 1 action keyword
		actionKeyword := a.actionKeywords[rand.Intn(len(a.actionKeywords))]
		contextKeywords = append(contextKeywords, actionKeyword)

		return AnalysisResult{
			Keywords:    contextKeywords,
			Confidence:  rand.Float64()*0.5 + 0.5, // 50-100% confidence
			IsFlagged:   true,
			Description: "Potential sensitive content detected",
		}
	}

	// If not flagged, just return context keywords
	if len(contextKeywords) > 0 {
		return AnalysisResult{
			Keywords:    contextKeywords,
			Confidence:  rand.Float64() * 0.3, // 0-30% confidence
			IsFlagged:   false,
			Description: "Normal activity detected",
		}
	}

	// Fallback for no context match
	return AnalysisResult{
		Keywords:    []string{"general", "activity"},
		Confidence:  0.1,
		IsFlagged:   false,
		Description: "No specific activity pattern detected",
	}
}
