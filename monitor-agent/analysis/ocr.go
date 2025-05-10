package analysis

import (
	"github.com/otiai10/gosseract/v2"
)

func (a *FakeAnalyzer) FileExtraction(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage(imagePath)
	text, err := client.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}
