package log

import (
	"fmt"
	"testing"
)

func TestLogShouldDropOldestLogsIfBufferIsFull(t *testing.T) {
	b := newLogBuffer(5)
	for i := range 10 {
		b.Ingest(Entry{Message: fmt.Sprintf("log %d", i)})
	}
	logged := []string{}
	b.AddHook("logged", func(log Entry) {
		logged = append(logged, log.Message)
	})
	equal(t, 5, len(logged), "should have 5 logs")
	equal(t, "log 5", logged[0], "should have log 5")
	equal(t, "log 6", logged[1], "should have log 6")
	equal(t, "log 7", logged[2], "should have log 7")
	equal(t, "log 8", logged[3], "should have log 8")
	equal(t, "log 9", logged[4], "should have log 9")
}

func TestLogDropShouldUpdateOffsets(t *testing.T) {
	b := newLogBuffer(5)
	logged := []string{}
	b.AddHook("logged", func(log Entry) {
		logged = append(logged, log.Message)
	})
	b.Ingest(Entry{Message: "before 1"})
	b.Ingest(Entry{Message: "before 2"})
	b.RemoveHook("logged")

	for i := range 10 {
		b.Ingest(Entry{Message: fmt.Sprintf("log %d", i)})
	}
	b.AddHook("logged", func(log Entry) {
		logged = append(logged, log.Message)
	})
	equal(t, 7, len(logged), "should have 7 logs")
	equal(t, "before 1", logged[0], "should have before 1")
	equal(t, "before 2", logged[1], "should have before 2")
	equal(t, "log 5", logged[2], "should have log 5")
	equal(t, "log 6", logged[3], "should have log 6")
	equal(t, "log 7", logged[4], "should have log 7")
	equal(t, "log 8", logged[5], "should have log 8")
	equal(t, "log 9", logged[6], "should have log 9")
}

func TestShouldNotRaiseIfRemoveNonExistentHook(_ *testing.T) {
	RemoveHook("nonExistentHook")
}
