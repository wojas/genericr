/*
Copyright 2019 The logr Authors.
Copyright 2020 The genericr Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package genericr implements github.com/go-logr/logr.Logger in a generic way
// that allows easy implementation of other logging backends.
package genericr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/go-logr/logr"
)

// Entry is a log entry that your adapter will receive for actual logging
type Entry struct {
	Level   int
	Prefix  []string
	Message string
	Error   error
	Fields  []interface{} // alternating key-value pairs
	// TODO: CallerDepth int
}

// String converts the entry to a string. The format is subject to change.
// TODO: Neater way to log values with newlines?
func (e Entry) String() string {
	var fieldsStr string
	if len(e.Fields) > 0 {
		fieldsStr = flatten(e.Fields...)
	}
	var errStr string
	if e.Error != nil {
		errStr = flatten("error", e.Error.Error()) + " "
	}
	return fmt.Sprintf("[%d] %s %q %s%s",
		e.Level, strings.Join(e.Prefix, "."), e.Message, errStr, fieldsStr)
}

// FieldsMap converts the fields to a map.
// This map is also compatible with logrus.Fields.
func (e Entry) FieldsMap() map[string]interface{} {
	return fieldsMap(e.Fields)
}

// LogFunc is your custom log backend
type LogFunc func(e Entry)

// New returns a logr.Logger which is implemented by your custom LogFunc.
func New(f LogFunc) Logger {
	log := Logger{
		f:         f,
		verbosity: 1000,
	}
	return log
}

type Logger struct {
	f         LogFunc
	level     int           // current verbosity level
	verbosity int           // max verbosity level that we log
	prefix    []string      // list of prefix names
	values    []interface{} // key-value pairs
	depth     int           // call stack depth to figure out caller info
}

// WithVerbosity returns a new instance with given max verbosity level
func (l Logger) WithVerbosity(level int) Logger {
	ll := l.clone()
	ll.verbosity = level
	return ll
}

func (l Logger) Info(msg string, kvList ...interface{}) {
	l.logMessage(nil, msg, kvList...)
}

func (l Logger) Enabled() bool {
	return l.verbosity >= l.level
}

func (l Logger) Error(err error, msg string, kvList ...interface{}) {
	l.logMessage(err, msg, kvList...)
}

func (l Logger) V(level int) logr.Logger {
	ll := l.clone()
	ll.level += level
	return ll
}

func (l Logger) WithName(name string) logr.Logger {
	ll := l.clone()
	ll.prefix = append(ll.prefix, name)
	return ll
}

func (l Logger) WithValues(kvList ...interface{}) logr.Logger {
	ll := l.clone()
	ll.values = append(ll.values, kvList...)
	return ll
}

func (l Logger) clone() Logger {
	out := l
	l.values = copySlice(l.values)
	n := len(l.prefix)
	if n > 0 {
		l.prefix = l.prefix[:n:n] // cap to force copy on append
	}
	return out
}

func (l Logger) logMessage(err error, msg string, kvList ...interface{}) {
	if !l.Enabled() {
		return
	}
	var out []interface{}
	if len(l.values) == 0 && len(kvList) > 0 {
		out = kvList
	} else if len(l.values) > 0 && len(kvList) == 0 {
		out = l.values
	} else {
		out = make([]interface{}, len(l.values)+len(kvList))
		copy(out, l.values)
		copy(out[len(l.values):], kvList)
	}
	l.f(Entry{
		Level:   l.level,
		Prefix:  l.prefix,
		Message: msg,
		Error:   err,
		Fields:  out,
		//CallerDepth: l.depth,
	})
}

var _ logr.Logger = Logger{}

// Helper functions below

func pretty(value interface{}) string {
	jb, err := json.Marshal(value)
	if err != nil {
		jb, _ = json.Marshal(fmt.Sprintf("%v", value))
	}
	return string(jb)
}

func copySlice(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))
	copy(out, in)
	return out
}

// flatten converts a key-value list to a friendly string
func flatten(kvList ...interface{}) string {
	vals := fieldsMap(kvList)
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf := bytes.Buffer{}
	for i, k := range keys {
		v := vals[k]
		if i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(pretty(k))
		buf.WriteString("=")
		buf.WriteString(pretty(v))
	}
	return buf.String()
}

// fieldsMap converts the fields to a map.
func fieldsMap(fields []interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(fields))
	for i := 0; i < len(fields); i += 2 {
		k, ok := fields[i].(string)
		if !ok {
			k = fmt.Sprintf("!(%#v)", fields[i])
		}
		var v interface{}
		if i+1 < len(fields) {
			v = fields[i+1]
		}
		m[k] = v
	}
	return m
}
