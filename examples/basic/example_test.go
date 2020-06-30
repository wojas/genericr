package logrus_test

import (
	"fmt"

	"github.com/wojas/genericr"
)

func ExampleBasic() {
	log := genericr.New(func(e genericr.Entry) {
		fmt.Println("LOG:", e.String())
	})

	log.V(0).Info("hello world at info level")
	log.WithName("some").WithName("component").WithValues("foo", 42).V(1).Info(
		"event", "extra", "someval1")
	log.V(1).Info("another level")
	log.WithName("some-component").WithValues("x", 123).V(1).Info("some event", "y", 42)

	// Output:
	// LOG: [0]  "hello world at info level"
	// LOG: [1] some.component "event" extra="someval1" foo=42
	// LOG: [1]  "another level"
}
