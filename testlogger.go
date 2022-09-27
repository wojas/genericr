package genericr

import (
	"github.com/go-logr/logr"
)

// TLogSink is the subset of the testing.TB interface we need to log with it
type TLogSink interface {
	Log(args ...interface{})
}

// NewForTesting returns a LogSink for given testing.T or B.
// Note that the source line reference will be incorrect in all messages
// written by testing.T. There is nothing we can do about that, the call depth
// is hardcoded in there.
func NewForTesting(t TLogSink) logr.Logger {
	var f LogFunc = func(e Entry) {
		t.Log(e.String())
	}
	sink := New(f)
	return logr.New(sink)
}
