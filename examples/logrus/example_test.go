package logrus_test

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wojas/genericr"
)

func ExampleLogrus() {
	root := logrus.New()
	root.SetLevel(logrus.TraceLevel)
	root.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})
	root.Out = os.Stdout

	var lf genericr.LogFunc = func(e genericr.Entry) {
		var l *logrus.Entry = root.WithField("component", e.Name)
		if e.Error != nil {
			l = l.WithError(e.Error)
		}
		if len(e.Fields) != 0 {
			l = l.WithFields(e.FieldsMap())
		}
		logrusLevel := logrus.Level(e.Level) + logrus.InfoLevel // 0 = info
		if logrusLevel < logrus.ErrorLevel {
			logrusLevel = logrus.ErrorLevel
		} else if logrusLevel > logrus.TraceLevel {
			logrusLevel = logrus.TraceLevel
		}
		l.Log(logrusLevel, e.Message)
	}
	log := genericr.New(lf)
	log.V(0).Info("hello world at info level")
	log.WithName("some").WithName("component").WithValues("foo", 42).V(1).Info(
		"event", "extra", "someval1")
	log.V(1).Info("debug level")
	log.V(2).Info("trace level")
	log.V(3).Info("also trace level")
	log.V(-1).Info("warn level")
	log.V(-2).Info("error level")
	log.V(-3).Info("also error level")

	// Output:
	// level=info msg="hello world at info level" component=
	// level=debug msg=event component=some.component extra=someval1 foo=42
	// level=debug msg="debug level" component=
	// level=trace msg="trace level" component=
	// level=trace msg="also trace level" component=
	// level=warning msg="warn level" component=
	// level=error msg="error level" component=
	// level=error msg="also error level" component=
}
