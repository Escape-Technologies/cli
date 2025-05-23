// Package log provides logging implementation with hooks to send logs to Escape Platform
package log

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// Entry is a log line
type Entry struct {
	Level   logrus.Level `json:"level"`
	Message string       `json:"message"`
}

type logBuffer struct {
	lock        sync.Mutex
	queue       []Entry
	bufferSize  int
	hooks       map[string]func(Entry)
	hooksOffset map[string]int
}

func newLogBuffer(bufferSize int) *logBuffer {
	return &logBuffer{
		bufferSize:  bufferSize,
		queue:       make([]Entry, 0, bufferSize),
		hooks:       map[string]func(Entry){},
		hooksOffset: map[string]int{},
	}
}

// Warning: This function is not thread safe,
// lock should be held when calling this function.
func (b *logBuffer) sync() {
	for name, callback := range b.hooks {
		offset, ok := b.hooksOffset[name]
		if !ok {
			offset = 0
		}
		for i := offset; i < len(b.queue); i++ {
			callback(b.queue[i])
		}
		b.hooksOffset[name] = len(b.queue)
	}
}

func (b *logBuffer) Ingest(log Entry) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if len(b.queue) >= b.bufferSize {
		b.queue = b.queue[1:]
		for name, offset := range b.hooksOffset {
			if offset > 0 {
				b.hooksOffset[name] = offset - 1
			} else {
				delete(b.hooksOffset, name)
			}
		}
	}
	b.queue = append(b.queue, log)
	b.sync()
}

func (b *logBuffer) AddHook(name string, callback func(Entry)) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.hooks[name] = callback
	b.sync()
}

func (b *logBuffer) RemoveHook(name string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	delete(b.hooks, name)
}
