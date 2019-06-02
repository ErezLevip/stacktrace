package stacktrace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
)

type StackTrace struct {
	Caller     string    `json:"caller"`
	StackTrace string    `json:"stack_trace"`
	reader     io.Reader `json:"-"`
}

const (
	CallStackFrameFormat  = "%s:%v:func:%s\n\t"
	CallStackCallerFormat = "%s:%v"
)

func NewStackTrace(skip, maxCallStackSize int) *StackTrace {
	frames := getFrames(skip, maxCallStackSize)
	return &StackTrace{
		Caller:     getCaller(frames),
		StackTrace: getTrace(frames),
	}
}

func getFrames(skip, maxCallStackSize int) []runtime.Frame {
	selectedFrames := make([]runtime.Frame, 0)
	programCounters := make([]uintptr, maxCallStackSize)
	n := runtime.Callers(3+skip, programCounters)
	if n <= 0 {
		return selectedFrames
	}

	frames := runtime.CallersFrames(programCounters[:n])
	more := true
	var current runtime.Frame
	for {
		if !more {
			break
		}
		current, more = frames.Next()
		selectedFrames = append(selectedFrames, current)
	}

	return selectedFrames
}

func getCaller(frames []runtime.Frame) string {
	return fmt.Sprintf(CallStackCallerFormat, formatTrace(frames[0].File), frames[0].Line)
}
func getTrace(frames []runtime.Frame) string {
	sb := &strings.Builder{}
	for _, f := range frames[1:] {
		sb.WriteString(fmt.Sprintf(CallStackFrameFormat, f.File, f.Line, strings.TrimLeft(f.Func.Name(), f.File)))
	}

	return sb.String()
}

func formatTrace(s string) string {
	sp := strings.Split(s, "/")
	return strings.Join(sp[len(sp)-2:], "/")
}

func marshal(t *StackTrace) (io.Reader, error) {
	buff := &bytes.Buffer{}
	return buff, json.NewEncoder(buff).Encode(t)
}

func (t *StackTrace) Read(p []byte) (n int, err error) {
	if t.reader == nil {
		r, err := marshal(t)
		if err != nil {
			return -1, err
		}
		t.reader = r
	}

	return t.reader.Read(p)
}

func (t *StackTrace) ToJson() (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
