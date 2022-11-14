package genericr_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-logr/logr"
	"github.com/wojas/genericr"
)

func TestNewForTesting(t *testing.T) {
	log := genericr.NewForTesting(t)
	log.Info("hello world", "a", 1)
	log.WithName("foo").WithValues("a", 42).WithName("bar").V(1).WithValues("b", 1).Info("hello world2", "x", 123, "y", 234)
	t.Log("Normal test log")
}

func TestLogger_Table(t *testing.T) {
	var last genericr.Entry
	var lf genericr.LogFunc = func(e genericr.Entry) {
		last = e
		t.Log(e.String())
	}
	sink := genericr.New(lf)

	tt := []struct {
		f    func()
		want string
	}{
		{
			func() {
				sink.Info(0, "hello world")
			},
			`[0]  "hello world"`,
		},
		{
			func() {
				sink.Info(0, "hello world", "a", 1)
			},
			`[0]  "hello world" a=1`,
		},
		{
			func() {
				sink.Info(4, "hello world", "a", 1)
			},
			`[4]  "hello world" a=1`,
		},
		{
			func() {
				sink.WithName("somename").Info(0, "hello world", "a", 1)
			},
			`[0] somename "hello world" a=1`,
		},
		{
			func() {
				sink.WithName("somename").WithName("sub").Info(0, "hello world", "a", 1)
			},
			`[0] somename.sub "hello world" a=1`,
		},
		{
			func() {
				sink.WithName("somename").WithName("sub").Info(1, "hello world", "a", 1, "b", 2)
			},
			`[1] somename.sub "hello world" a=1 b=2`,
		},
		{
			func() {
				sink.WithValues("x", "yz").WithName("somename").WithName("sub").Info(1, "hello world", "a", 1, "b", 2)
			},
			`[1] somename.sub "hello world" a=1 b=2 x="yz"`,
		},
		{
			func() {
				// Odd values by mistake, does not corrupt later calls
				sink.WithValues("x", "yz", "z").Info(0, "hello world", "a", 1, "b", 2)
			},
			`[0]  "hello world" a=1 b=2 x="yz" z=null`,
		},
		{
			func() {
				sink.WithVerbosity(1).Info(0, "hello world", "a", 1)
			},
			`[0]  "hello world" a=1`,
		},
		{
			func() {
				sink.Info(01, "first")
				sink.WithVerbosity(1).Info(1, "hello world", "a", 1)
			},
			`[1]  "hello world" a=1`,
		},
		{
			func() {
				sink.Info(0, "first")
				sink.WithVerbosity(1).Info(0, "hello world", "a", 1)
			},
			`[0]  "hello world" a=1`,
		},
		{
			func() {
				sink.Info(0, "wrong params", "a")
			},
			`[0]  "wrong params" a=null`,
		},
		{
			func() {
				sink.Info(0, "wrong params", 42)
			},
			`[0]  "wrong params" "!(42)"=null`,
		},
		{
			func() {
				sink.Error(fmt.Errorf("some error"), "help")
			},
			`[0]  "help" error="some error"`,
		},
		{
			f: func() {
				sink.WithValues(
					"int", 42,
					"string", "foo",
					"bytes", []byte("foo"),
					"float", 3.14,
					"struct", struct {
						A string
						B int
					}{"foo", 12},
					"map", map[string]int{
						"foo": 12,
					},
					"nilval", nil,
					"err", errors.New("oops"),
					"stringslice", []string{"a", "b"},
				).Info(0, "types")
			},
			want: `[0]  "types" bytes="66 6f 6f" err="oops" float=3.14 int=42 map={"foo":12} nilval=null string="foo" stringslice=["a","b"] struct={"A":"foo","B":12}`,
		},
	}

	for i, row := range tt {
		row.f()
		s := last.String()
		if s != row.want {
			t.Errorf("row %d:\n  got:  `%s`\n  want: `%s`", i, s, row.want)
		}
	}
}

func TestLogger_Caller(t *testing.T) {
	var last genericr.Entry
	var lf genericr.LogFunc = func(e genericr.Entry) {
		last = e
		t.Log(e.String())
		t.Log(runtime.Caller(e.CallerDepth)) // should log caller_test
	}
	sink := genericr.New(lf).WithCaller(true)
	logSomethingFromOtherFile(logr.New(sink))

	_, fname := filepath.Split(last.Caller.File)
	if fname != "caller_test.go" {
		t.Errorf("Caller: expected 'caller_test.go', got %q (full: %s:%d)",
			fname, last.Caller.File, last.Caller.Line)
	}
	if last.CallerDepth != 4 {
		t.Errorf("Caller depth: expected 4, got %d", last.CallerDepth)
	}
}

type wrappedLogger struct {
	log logr.Logger
}

func (wl wrappedLogger) Info(msg string, keysAndValues ...interface{}) {
	wl.log.Info(msg, keysAndValues...)
}

func TestLogger_WithCallDepth(t *testing.T) {
	var last genericr.Entry
	var lf genericr.LogFunc = func(e genericr.Entry) {
		last = e
		t.Log(e.String())
		t.Log(runtime.Caller(e.CallerDepth)) // should log caller_test
	}
	sink := genericr.New(lf).WithCaller(true).WithCallDepth(1)
	logger := logr.New(sink)
	wlogger := wrappedLogger{logger}

	logSomethingFromOtherFile(wlogger)

	_, fname := filepath.Split(last.Caller.File)
	if fname != "caller_test.go" {
		t.Errorf("Caller: expected 'caller_test.go', got %q (full: %s:%d)",
			fname, last.Caller.File, last.Caller.Line)
	}
	if last.CallerDepth != 5 {
		t.Errorf("Caller depth: expected 5, got %d", last.CallerDepth)
	}
}

func BenchmarkLogger_basic(b *testing.B) {
	foo := 0
	var lf genericr.LogFunc = func(e genericr.Entry) {
		foo += e.Level // just to prevent it from being optimized away
	}
	log := genericr.New(lf)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(0, "hello")
	}
}

func BenchmarkLogger_basic_with_caller(b *testing.B) {
	foo := 0
	var lf genericr.LogFunc = func(e genericr.Entry) {
		foo += e.Level // just to prevent it from being optimized away
	}
	log := genericr.New(lf).WithCaller(true)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(0, "hello")
	}
}

func BenchmarkLogger_2vars(b *testing.B) {
	foo := 0
	var lf genericr.LogFunc = func(e genericr.Entry) {
		foo += e.Level // just to prevent it from being optimized away
	}
	log := genericr.New(lf)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(0, "hello", "a", 1, "b", 2)
	}
}

func BenchmarkLogger_clone(b *testing.B) {
	foo := 0
	var lf genericr.LogFunc = func(e genericr.Entry) {
		foo += e.Level // just to prevent it from being optimized away
	}
	log := genericr.New(lf)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(0, "hello")
	}
}

func BenchmarkLogger_complicated(b *testing.B) {
	foo := 0
	var lf genericr.LogFunc = func(e genericr.Entry) {
		foo += e.Level // just to prevent it from being optimized away
	}
	log := genericr.New(lf)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.WithName("bench").WithValues("x", 123).Info(0, "hello", "a", 1, "b", 2)
	}
}

func BenchmarkLogger_complicated_precalculated(b *testing.B) {
	foo := 0
	var lf genericr.LogFunc = func(e genericr.Entry) {
		foo += e.Level // just to prevent it from being optimized away
	}
	log := genericr.New(lf)
	log2 := log.WithName("bench").WithValues("x", 123)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log2.Info(0, "hello", "a", 1, "b", 2)
	}
}

func BenchmarkLogger_2vars_tostring(b *testing.B) {
	foo := ""
	var lf genericr.LogFunc = func(e genericr.Entry) {
		foo = e.String()
	}
	log := genericr.New(lf)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(0, "hello", "a", 1, "b", 2)
	}
	_ = foo
}
