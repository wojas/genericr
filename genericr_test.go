package genericr_test

import (
	"fmt"
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
			`[0]  "hello world" `,
		},
		{
			func() {
				log.Info("hello world", "a", 1)
			},
			`[0]  "hello world" "a"=1`,
		},
		{
			func() {
				log.V(4).Info("hello world", "a", 1)
			},
			`[4]  "hello world" "a"=1`,
		},
		{
			func() {
				log.WithName("somename").Info("hello world", "a", 1)
			},
			`[0] somename "hello world" "a"=1`,
		},
		{
			func() {
				log.WithName("somename").WithName("sub").Info("hello world", "a", 1)
			},
			`[0] somename.sub "hello world" "a"=1`,
		},
		{
			func() {
				log.WithName("somename").V(1).WithName("sub").Info("hello world", "a", 1, "b", 2)
			},
			`[1] somename.sub "hello world" "a"=1 "b"=2`,
		},
		{
			func() {
				log.WithValues("x", "yz").WithName("somename").V(1).WithName("sub").Info("hello world", "a", 1, "b", 2)
			},
			`[1] somename.sub "hello world" "a"=1 "b"=2 "x"="yz"`,
		},
		{
			func() {
				log.WithVerbosity(1).Info("hello world", "a", 1)
			},
			`[0]  "hello world" "a"=1`,
		},
		{
			func() {
				log.Info("first")
				log.WithVerbosity(1).V(1).Info("hello world", "a", 1)
			},
			`[1]  "hello world" "a"=1`,
		},
		{
			func() {
				log.Info("first")
				log.WithVerbosity(1).V(2).Info("hello world", "a", 1)
			},
			`[0]  "first" `,
		},
		{
			func() {
				log.Info("wrong params", "a")
			},
			`[0]  "wrong params" "a"=null`,
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
			`[0]  "help" "error"="some error" `,
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

func BenchmarkLogger_2vars_string(b *testing.B) {
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
