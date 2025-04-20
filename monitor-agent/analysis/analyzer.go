package analysis

import (
	"math/rand"
	"strings"
	"time"
)

type Analyzer struct {
	sensitiveKeywords []string
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		sensitiveKeywords: []string{
			"password",
			"credit card",
			"social security",
			"confidential",
			"private",
		},
	}
}

func (a *Analyzer) Analyze(content string) (string, float64) {
	// Mock implementation
	// In a real implementation, this would use more sophisticated analysis
	content = strings.ToLower(content)

	// Check for sensitive keywords
	for _, keyword := range a.sensitiveKeywords {
		if strings.Contains(content, keyword) {
			return "Potential sensitive content detected", 0.8 + rand.Float64()*0.2
		}
	}

	// Random analysis for demonstration
	rand.Seed(time.Now().UnixNano())
	if rand.Float64() < 0.3 {
		return "Content flagged for review", 0.5 + rand.Float64()*0.3
	}

	return "Content appears normal", 0.1 + rand.Float64()*0.2
}
