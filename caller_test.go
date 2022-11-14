package genericr_test

type infoLogger interface {
	Info(msg string, keysAndValues ...interface{})
}

// This is indirectly tested by calling this function
func logSomethingFromOtherFile(log infoLogger) {
	log.Info("test caller")
}
