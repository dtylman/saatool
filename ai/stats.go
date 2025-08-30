package ai

import (
	"sync"
	"time"
)

type statItem struct {
	count     int
	startTime time.Time
}

// TranslationStatistics provides statistics about the translation process
type TranslationStatistics struct {
	paragraphsInProgress map[string]statItem
	mutex                sync.Mutex
	averageWordTime      time.Duration
	totalWords           int
	totalTime            time.Duration
}

// NewTranslationStatistics creates a new instance of TranslationStatistics
func NewTranslationStatistics() *TranslationStatistics {
	return &TranslationStatistics{
		paragraphsInProgress: make(map[string]statItem),
	}
}

// Started marks a paragraph as started
func (t *TranslationStatistics) Started(paragraphID string, wordCount int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.paragraphsInProgress[paragraphID] = statItem{
		count:     wordCount,
		startTime: time.Now(),
	}
	t.totalWords += wordCount
}

// Completed marks a paragraph as completed
func (t *TranslationStatistics) Completed(paragraphID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	item, exists := t.paragraphsInProgress[paragraphID]
	if !exists {
		return
	}
	duration := time.Since(item.startTime)
	t.totalTime += duration
	if t.totalWords > 0 {
		t.averageWordTime = t.totalTime / time.Duration(t.totalWords)
	} else {
		t.averageWordTime = 0
	}
	delete(t.paragraphsInProgress, paragraphID)
}

// NextETA estimates the time remaining for the next paragraph to complete
func (t *TranslationStatistics) NextETA() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	oldest := statItem{}
	for _, item := range t.paragraphsInProgress {
		if oldest.startTime.IsZero() || item.startTime.Before(oldest.startTime) {
			oldest = item
		}
	}
	if oldest.startTime.IsZero() {
		return 0
	}

	expectedDuration := time.Duration(oldest.count) * t.averageWordTime
	elapsed := time.Since(oldest.startTime)
	remaining := expectedDuration - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// InProgressCount returns the number of paragraphs currently being translated
func (t *TranslationStatistics) InProgressCount() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return len(t.paragraphsInProgress)
}
