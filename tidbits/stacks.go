package tidbits

import (
	"encoding"
	"encoding/json"
	"fmt"
	"log/slog"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const StackAttrName = "stack"

type StackValue struct {
	skipToFirstPanic bool
	stack            []uintptr
	msg              string
}

var _ json.Marshaler = &StackValue{}
var _ encoding.TextMarshaler = &StackValue{}

func StackTraceAttr(skipToFirstPanic bool, msg string) slog.Attr {
	return slog.Any(StackAttrName, NewStackValue(2, skipToFirstPanic, msg))
}

func NewStackValue(skipFrames int, skipToFirstPanic bool, msg any) *StackValue {
	stack := make([]uintptr, 128)
	num := runtime.Callers(skipFrames, stack)

	return &StackValue{skipToFirstPanic: skipToFirstPanic, stack: stack[:num], msg: PanicMsgToString(msg)}
}

func PanicMsgToString(msg interface{}) string {
	if msg == nil {
		return "recovered from panic"
	}
	stringer, ok := msg.(fmt.Stringer)
	if ok {
		return stringer.String()
	}
	err, ok := msg.(error)
	if ok {
		return err.Error()
	}
	return reflect.ValueOf(msg).String()
}

type StackElement struct {
	Msg string `json:"panic_msg,omitempty"`
	Fl  string `json:"fl,omitempty"`
	Fn  string `json:"fn,omitempty"`
}

func (s *StackValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.JSONStack())
}

// JSONStack creates a nice stack trace, skipping all the deferred frames after the first panic() call.
// This method returns the list of structures that can be nicely reflected into JSON.
func (s *StackValue) JSONStack() []StackElement {
	// Create the stack trace
	stackElements := make([]StackElement, 0, 20)
	stackElements = append(stackElements, StackElement{Msg: s.msg})

	panicsToSkip := 0
	if s.skipToFirstPanic {
		panicsToSkip = s.countPanics()
	}

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is the runtime frame which
	// adds noise, since it always starts in the runtime.
	frames := runtime.CallersFrames(s.stack)
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		filePath, line, label := s.parseFrame(frame)

		if panicsToSkip > 0 && strings.HasPrefix(filePath, "runtime/panic") && label == "gopanic" {
			panicsToSkip -= 1
			continue
		}
		if panicsToSkip > 0 {
			continue
		}

		stackElements = append(stackElements, StackElement{
			Fl: filePath + ":" + strconv.Itoa(line),
			Fn: label,
		})
	}
	return stackElements
}

// MarshalText creates a nice stack trace, skipping all the deferred frames after the first panic() call.
// This method returns a human-readable multi-line string.
func (s *StackValue) MarshalText() (text []byte, err error) {
	// Create the stack trace
	frames := runtime.CallersFrames(s.stack)

	panicsToSkip := 0
	if s.skipToFirstPanic {
		panicsToSkip = s.countPanics()
	}

	var res string
	res += s.msg + "\n"

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is the runtime frame which
	// adds noise, since it always starts in the runtime.
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		filePath, line, label := s.parseFrame(frame)

		if panicsToSkip > 0 && strings.HasPrefix(filePath, "runtime/panic") && label == "gopanic" {
			panicsToSkip -= 1
			continue
		}
		if panicsToSkip > 0 {
			continue
		}

		res += filePath + ":" + strconv.Itoa(line) + " " + label + "\n"
	}

	return []byte(res), nil
}

// The default stack trace contains the build environment full path as the first part of the file name.
// This adds no information to the stack trace and exposes the building environment,
// so process the stack trace to remove the building environment path.
func (s *StackValue) parseFrame(frame runtime.Frame) (string, int, string) {
	// Example:
	// frame.Function = github.com/Cyberax/slog-tidbits/tidbits.StackTraceAttr
	// frame.Line = 18
	// frame.File = /Users/cyberax/bricks/slog-tidbits/tidbits/stacks.go
	fname := path.Base(frame.File)
	dotIdx := strings.LastIndex(frame.Function, ".")

	packagePath := fname
	funcName := frame.Function
	if dotIdx != -1 {
		packagePath = frame.Function[:dotIdx] + "/" + fname
		funcName = frame.Function[dotIdx+1:]
	}

	// github.com/Cyberax/slog-tidbits/tidbits/stacks.go, 11, NewStackValue
	return packagePath, frame.Line, funcName
}

// Count the number of go panic() calls in the stack trace
func (s *StackValue) countPanics() int {
	frames := runtime.CallersFrames(s.stack)
	panics := 0
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		filePath, _, label := s.parseFrame(frame)
		if strings.HasPrefix(filePath, "runtime/panic") && label == "gopanic" {
			panics += 1
		}
	}
	return panics
}
