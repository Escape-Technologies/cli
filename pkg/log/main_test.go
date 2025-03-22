package log

import (
	"testing"
)

func equal(t *testing.T, expected, actual any, msg string) {
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func TestWorkflow(t *testing.T) {
	Info("0")
	Info("1")

	buf1 := []string{}
	AddHook("test1", func(log LogItem) { buf1 = append(buf1, log.Message) })
	buf2 := []string{}
	AddHook("test2", func(log LogItem) { buf2 = append(buf2, log.Message) })

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

	AddHook("test1", func(log LogItem) { buf1 = append(buf1, log.Message) })
	equal(t, 4, len(buf1), "should have four logs in buf1")
	equal(t, "3", buf1[3], "should have fourth log in buf1")
}
