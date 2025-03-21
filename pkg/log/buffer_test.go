package log

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func equal(t *testing.T, expected, actual interface{}, msg string) {
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func TestLogBuffer(t *testing.T) {
	bufferSize := 10
	buffer := newLogBuffer(bufferSize)
	
	testMessages := []string{"test 1", "test 2", "test 3", "test 4", "test 5"}
	for _, msg := range testMessages {
		buffer.AddLog(msg)
	}

	firstLog := buffer.GetLog()
	equal(t, testMessages[0], firstLog, "should get first message")
	
	secondLog := buffer.GetLog()
	equal(t, testMessages[1], secondLog, "should get second message")

	remainingLogs := buffer.GetLogs()
	equal(t, 3, len(remainingLogs), "should have three remaining messages")
	equal(t, testMessages[2], remainingLogs[0], "should get third message")
	equal(t, testMessages[3], remainingLogs[1], "should get fourth message")
	equal(t, testMessages[4], remainingLogs[2], "should get fifth message")

	emptyLog := buffer.GetLog()
	equal(t, "", emptyLog, "should return empty string when no logs available")

	buffer.Close()
}

func TestGlobalBufferWithHook(t *testing.T) {
	assert.NotNil(t, globalBuffer, "global buffer should be initialized")
	
	Info("message 1")
	Warn("message 2")
	Error("message 3")
	Info("message 4")
	Warn("message 5")

	firstLog := globalBuffer.GetLog()
	equal(t, "message 1", firstLog, "should get first message")

	remainingLogs := globalBuffer.GetLogs()
	equal(t, 4, len(remainingLogs), "should have four remaining messages")
	equal(t, "message 2", remainingLogs[0], "should get second message")
	equal(t, "message 3", remainingLogs[1], "should get third message")
	equal(t, "message 4", remainingLogs[2], "should get fourth message")
	equal(t, "message 5", remainingLogs[3], "should get fifth message")	

	emptyLog := globalBuffer.GetLog()
	equal(t, "", emptyLog, "should return empty string when buffer is empty")
}

func TestShouldNotRaiseIfRemoveNonExistentHook(t *testing.T) {
	RemoveHook("nonExistentHook")
}

func TestWorkflow(t *testing.T) {
	assert.NotNil(t, globalBuffer, "global buffer should be initialized")
	
	Info("0")
	Info("1")

	buf1 := []string{}
	AddHook("test1", func(_ logrus.Level, message string) {buf1 = append(buf1, message)})
	buf2 := []string{}
	AddHook("test2", func(_ logrus.Level, message string) {buf2 = append(buf2, message)})

	// Logs should be sent in order to hooks
	equal(t, 2, len(buf1), "should have two logs in buf1")
	equal(t, "0", buf1[0], "should have first log in buf1")
	equal(t, "1", buf1[1], "should have second log in buf1")
	equal(t, 2, len(buf2), "should have two logs in buf2")
	equal(t, "0", buf2[0], "should have first log in buf2")
	equal(t, "1", buf2[1], "should have second log in buf2")

    Info("2")
	equal(t, 3, len(buf1), "should have three logs in buf1")
	equal(t, "2", buf1[2], "should have third log in buf1")
	equal(t, 3, len(buf2), "should have three logs in buf2")
	equal(t, "2", buf2[2], "should have third log in buf2")

	RemoveHook("test1")

	Info("3")
	equal(t, 3, len(buf1), "should have three logs in buf1")
	equal(t, 4, len(buf2), "should have four logs in buf2")
	equal(t, "3", buf2[3], "should have fourth log in buf2")

	AddHook("test1", func(_ logrus.Level, message string) {buf1 = append(buf1, message)})
	equal(t, 4, len(buf1), "should have four logs in buf1")
	equal(t, "3", buf1[3], "should have fourth log in buf1")
}
