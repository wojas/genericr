package genericr_test

import "github.com/go-logr/logr"

// This is indirectly tested by calling this function
func logSomethingFromOtherFile(log logr.Logger) {
	log.Info("test caller")
}
