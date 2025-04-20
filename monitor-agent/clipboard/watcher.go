package clipboard

import (
	"github.com/atotto/clipboard"
)

type Watcher struct {
	lastContent string
}

func NewWatcher() *Watcher {
	return &Watcher{}
}

func (w *Watcher) GetClipboardContent() string {
	content, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	return content
}

func (w *Watcher) HasChanged() bool {
	current := w.GetClipboardContent()
	if current != w.lastContent {
		w.lastContent = current
		return true
	}
	return false
}
