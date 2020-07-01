package genericr_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

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
	log := genericr.New(lf)

	tt := []struct {
		f    func()
		want string
	}{
		{
			func() {
				log.Info("hello world")
			},
			`[0]  "hello world"`,
		},
		{
			func() {
				log.Info("hello world", "a", 1)
			},
			`[0]  "hello world" a=1`,
		},
		{
			func() {
				log.V(4).Info("hello world", "a", 1)
			},
			`[4]  "hello world" a=1`,
		},
		{
			func() {
				log.WithName("somename").Info("hello world", "a", 1)
			},
			`[0] somename "hello world" a=1`,
		},
		{
			func() {
				log.WithName("somename").WithName("sub").Info("hello world", "a", 1)
			},
			`[0] somename.sub "hello world" a=1`,
		},
		{
			func() {
				log.WithName("somename").V(1).WithName("sub").Info("hello world", "a", 1, "b", 2)
			},
			`[1] somename.sub "hello world" a=1 b=2`,
		},
		{
			func() {
				log.WithValues("x", "yz").WithName("somename").V(1).WithName("sub").Info("hello world", "a", 1, "b", 2)
			},
			`[1] somename.sub "hello world" a=1 b=2 x="yz"`,
		},
		{
			func() {
				// Odd values by mistake, does not corrupt later calls
				log.WithValues("x", "yz", "z").Info("hello world", "a", 1, "b", 2)
			},
			`[0]  "hello world" a=1 b=2 x="yz" z=null`,
		},
		{
			func() {
				log.WithVerbosity(1).Info("hello world", "a", 1)
			},
			`[0]  "hello world" a=1`,
		},
		{
			func() {
				log.Info("first")
				log.WithVerbosity(1).V(1).Info("hello world", "a", 1)
			},
			`[1]  "hello world" a=1`,
		},
		{
			func() {
				log.Info("first")
				log.WithVerbosity(1).V(2).Info("hello world", "a", 1)
			},
			`[0]  "first"`,
		},
		{
			func() {
				log.Info("wrong params", "a")
			},
			`[0]  "wrong params" a=null`,
		},
		{
			func() {
				log.Info("wrong params", 42)
			},
			`[0]  "wrong params" "!(42)"=null`,
		},
		{
			func() {
				log.Error(fmt.Errorf("some error"), "help")
			},
			`[0]  "help" error="some error"`,
		},
		{
			f: func() {
				log.WithValues(
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
				).Info("types")
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
	log := genericr.New(lf).WithCaller(true)
	logSomethingFromOtherFile(log)

	_, fname := filepath.Split(last.Caller.File)
	if fname != "caller_test.go" {
		t.Errorf("Caller: expected 'caller_test.go', got %q (full: %s:%d)",
			fname, last.Caller.File, last.Caller.Line)
	}
	if last.CallerDepth != 3 {
		t.Errorf("Caller depth: expected 3, got %d", last.CallerDepth)
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
		log.Info("hello")
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
		log.Info("hello")
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
		log.Info("hello", "a", 1, "b", 2)
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
		log.V(1).Info("hello")
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
		log.V(1).WithName("bench").WithValues("x", 123).Info("hello", "a", 1, "b", 2)
	}
}

func BenchmarkLogger_complicated_precalculated(b *testing.B) {
	foo := 0
	var lf genericr.LogFunc = func(e genericr.Entry) {
		foo += e.Level // just to prevent it from being optimized away
	}
	log := genericr.New(lf)
	log2 := log.V(1).WithName("bench").WithValues("x", 123)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log2.Info("hello", "a", 1, "b", 2)
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
		log.Info("hello", "a", 1, "b", 2)
	}
	_ = foo
}
