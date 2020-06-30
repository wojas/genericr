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
	"regexp"
	"sort"
	"strings"

	"github.com/go-logr/logr"
)

// Entry is a log entry that your adapter will receive for actual logging
type Entry struct {
	Level     int           // level at which this was logged
	Name      string        // name parts joined with '.'
	NameParts []string      // individual name segments
	Message   string        // message as send to log call
	Error     error         // error if .Error() was called
	Fields    []interface{} // alternating key-value pairs
	// TODO: CallerDepth int
}

// String converts the entry to a string.
// The output format is subject to change! Implement your own conversion if
// you need to parse these logs later!
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
		e.Level, strings.Join(e.NameParts, "."), e.Message, errStr, fieldsStr)
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

// Logger is a generic logger that implements the logr.Logger interface and
// calls a function of type LogFunc for every log message received.
type Logger struct {
	f         LogFunc
	level     int           // current verbosity level
	verbosity int           // max verbosity level that we log
	nameParts []string      // list of names
	name      string        // nameParts joined by '.' for performance
	values    []interface{} // key-value pairs
	depth     int           // call stack depth to figure out caller info
}

// WithVerbosity returns a new instance with given max verbosity level
func (l Logger) WithVerbosity(level int) Logger {
	l.verbosity = level
	return l
}

func (l Logger) Info(msg string, kvList ...interface{}) {
	l.logMessage(nil, msg, kvList)
}

func (l Logger) Enabled() bool {
	return l.verbosity >= l.level
}

func (l Logger) Error(err error, msg string, kvList ...interface{}) {
	l.logMessage(err, msg, kvList)
}

func (l Logger) V(level int) logr.Logger {
	l.level += level
	return l
}

func (l Logger) WithName(name string) logr.Logger {
	// We keep both a list of parts for full flexibility, and a pre-joined string
	// for performance. We assume that this method is called far less often
	// than that actual logging is done.
	if len(l.nameParts) == 0 {
		l.nameParts = []string{name}
		l.name = name
	} else {
		n := len(l.nameParts)
		l.nameParts = append(l.nameParts[:n:n], name) // triple-slice to force copy
		l.name += "." + name
	}
	return l
}

func (l Logger) WithValues(kvList ...interface{}) logr.Logger {
	if len(kvList) == 0 {
		return l
	}
	if len(kvList)%2 == 1 {
		// Ensure an odd number of items here does not corrupt the list
		kvList = append(kvList, nil)
	}
	if len(l.values) == 0 {
		l.values = kvList
	} else {
		n := len(l.values)
		l.values = append(l.values[:n:n], kvList...) // triple-slice to force copy
	}
	return l
}

func (l Logger) logMessage(err error, msg string, kvList []interface{}) {
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
		Level:     l.level,
		Name:      l.name,
		NameParts: l.nameParts,
		Message:   msg,
		Error:     err,
		Fields:    out,
		//CallerDepth: l.depth,
	})
}

var _ logr.Logger = Logger{}

// Helper functions below

var safeString = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func prettyKey(v string) string {
	if safeString.MatchString(v) {
		return v
	} else {
		return pretty(v)
	}
}

func pretty(value interface{}) string {
	switch v := value.(type) {
	case error:
		value = v.Error()
	case []byte:
		return fmt.Sprintf(`"% x"`, v)
	}
	jb, err := json.Marshal(value)
	if err != nil {
		jb, _ = json.Marshal(fmt.Sprintf("%q", value))
	}
	return string(jb)
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
		buf.WriteString(prettyKey(k))
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
