package analysis

import (
	"math/rand"
	"time"
)

type AnalysisResult struct {
	Keywords    []string
	Confidence  float64
	IsFlagged   bool
	Description string
}

type FakeAnalyzer struct {
	keywords []string
}

func NewFakeAnalyzer() *FakeAnalyzer {
	return &FakeAnalyzer{
		keywords: []string{
			"confidential",
			"password",
			"secret",
			"private",
			"classified",
		},
	}
}

func (a *FakeAnalyzer) AnalyzeScreenshot(imagePath string) AnalysisResult {
	// Simulate random analysis
	rand.Seed(time.Now().UnixNano())

	// 30% chance of finding something
	if rand.Float64() < 0.3 {
		// Randomly select 1-3 keywords
		numKeywords := rand.Intn(3) + 1
		foundKeywords := make([]string, numKeywords)
		for i := 0; i < numKeywords; i++ {
			foundKeywords[i] = a.keywords[rand.Intn(len(a.keywords))]
		}

		return AnalysisResult{
			Keywords:    foundKeywords,
			Confidence:  rand.Float64()*0.5 + 0.5, // 50-100% confidence
			IsFlagged:   true,
			Description: "Potential sensitive content detected",
		}
	}

	return AnalysisResult{
		Keywords:    []string{},
		Confidence:  0.0,
		IsFlagged:   false,
		Description: "No sensitive content detected",
	}
}
