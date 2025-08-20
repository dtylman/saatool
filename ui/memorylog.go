package ui

import (
	"io"
	"log"
	"sync"
)

var (
	MemoryLog = newMemoryLogBuffer()
)

type memoryLogBuffer struct {
	mutex    sync.Mutex
	messages []string
	maxSize  int
}

func newMemoryLogBuffer() *memoryLogBuffer {
	return &memoryLogBuffer{
		messages: make([]string, 0),
		maxSize:  500, // default max size
	}
}

func (ml *memoryLogBuffer) Write(p []byte) (n int, err error) {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	ml.messages = append(ml.messages, string(p))
	if len(ml.messages) > ml.maxSize {
		ml.messages = ml.messages[len(ml.messages)-ml.maxSize:]
	}
	return len(p), nil
}

func (ml *memoryLogBuffer) Init() {
	log.Printf("initializing memory log buffer with max size %d", ml.maxSize)
	teeWriter := io.MultiWriter(ml, log.Writer())
	log.SetOutput(teeWriter)
}

// GetMessages retrieves the log messages from the memory buffer.
func (ml *memoryLogBuffer) Len() int {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()
	return len(ml.messages)
}

// GetMessages retrieves the log messages from the memory buffer.
func (ml *memoryLogBuffer) GetMessages() []string {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()
	return append([]string(nil), ml.messages...) // return a copy of the messages
}
