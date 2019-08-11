package stacktrace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/json-iterator/go"
	"io"
	"runtime"
	"strings"
	"sync"
)

type StackTrace struct {
	Caller     string    `json:"caller"`
	StackTrace string    `json:"stack_trace"`
	reader     io.Reader `json:"-"`
}

const (
	callStackFrameFormat  = "%s:%v:func:%s\n\t"
	callStackCallerFormat = "%s:%v"
	maxFramesSize         = 10
	skipBase              = 3
)

func NewStackTrace(skip int) *StackTrace {
	frames := getFrames(skip + skipBase)
	return &StackTrace{
		Caller:     getCaller(frames),
		StackTrace: getTrace(frames),
	}
}

var framesPool *sync.Pool
var ptrsPool *sync.Pool

func init() {
	framesPool = &sync.Pool{
		New: func() interface{} { return make([]runtime.Frame, maxFramesSize) },
	}
	ptrsPool = &sync.Pool{
		New: func() interface{} { return make([]uintptr, maxFramesSize) },
	}
}

func getFrames(skip int) []runtime.Frame {
	i := 0
	selectedFrames := framesPool.Get().([]runtime.Frame)
	programCounters := ptrsPool.Get().([]uintptr)

	defer func() {
		framesPool.Put(selectedFrames)
		ptrsPool.Put(programCounters)
	}()

	n := runtime.Callers(skip, programCounters)
	if n <= 0 {
		return selectedFrames
	}

	frames := runtime.CallersFrames(programCounters[:n])

	more := true
	var current runtime.Frame
	for {
		if !more || i >= maxFramesSize {
			break
		}

		current, more = frames.Next()
		selectedFrames[i] = current
		i++
	}

	return selectedFrames[:i]
}

func getCaller(frames []runtime.Frame) string {
	return fmt.Sprintf(callStackCallerFormat, formatTrace(frames[0].File), frames[0].Line)
}
func getTrace(frames []runtime.Frame) string {
	sb := &strings.Builder{}
	for _, f := range frames[1:] {
		sb.WriteString(fmt.Sprintf(callStackFrameFormat, f.File, f.Line, strings.TrimLeft(f.Func.Name(), f.File)))
	}

	return sb.String()
}

func formatTrace(s string) string {
	sp := strings.Split(s, "/")
	if len(sp) < 2 {
		return s
	}
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

func (t *StackTrace) ToJson() (io.Reader, error) {
	buff := new(bytes.Buffer)
	if err := jsoniter.NewEncoder(buff).Encode(t); err != nil {
		return nil, err
	}
	return buff, nil
}

func (t *StackTrace) ToJsonString() (string, error) {
	b, err := jsoniter.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
