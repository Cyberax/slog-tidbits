package tidbits

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path"
	"slices"
	"sync"
	"time"
)

var ColorReset = "\033[0m"
var ColorRed = "\033[31m"
var ColorGreen = "\033[32m"
var ColorYellow = "\033[33m"
var ColorBlue = "\033[34m"
var Purple = "\033[35m"
var ColorCyan = "\033[36m"
var ColorGray = "\033[37m"
var ColorWhite = "\033[97m"

var BlackBackground = "\033[40m"
var RedBackground = "\033[41m"
var GreenBackground = "\033[42m"
var YellowBackground = "\033[43m"
var BlueBackground = "\033[44m"
var PurpleBackground = "\033[45m"
var CyanBackground = "\033[46m"
var GrayBackground = "\033[47m"
var WhiteBackground = "\033[107m"

type PrettySink struct {
	handler   *slog.JSONHandler
	separator string
	colorize  bool

	mtx      sync.Mutex
	delegate io.Writer
	buf      *bytes.Buffer
}

var _ io.Writer = &PrettySink{}

func NewPrettySink(delegate io.Writer, lvl slog.Level, colorize bool) *PrettySink {
	res := &PrettySink{
		separator: "  ",
		colorize:  colorize,
		delegate:  delegate,
		buf:       bytes.NewBuffer(nil),
	}
	res.handler = slog.NewJSONHandler(res, &slog.HandlerOptions{AddSource: true, Level: lvl})
	return res
}

func (p *PrettySink) GetHandler() *slog.JSONHandler {
	return p.handler
}

func (p *PrettySink) Write(data []byte) (n int, err error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	curRead := 0
	for {
		// We need to split the data into lines and then parse each line separately
		idx := slices.Index(data, '\n')
		if idx == -1 {
			p.buf.Write(data)
			break
		}
		curStr := data[:idx]
		data = data[idx+1:]
		curRead += len(curStr)

		if p.buf.Len() > 0 {
			// We have the beginning of the string in the buffer, so we need to append the current string to it
			p.buf.Write(curStr)
			err = p.prettyPrint(p.buf.Bytes())
			p.buf.Reset()
			if err != nil {
				return curRead, err
			}
		} else {
			// Optimization: this is a complete string, so no need to use the buffer
			err = p.prettyPrint(curStr)
			if err != nil {
				return curRead, err
			}
		}
	}
	return len(data), nil
}

func valAsStr(data map[string]any, key string) string {
	val, ok := data[key]
	if !ok {
		return ""
	}
	strVal, ok := val.(string)
	if ok {
		return strVal
	}
	return fmt.Sprint(val)
}

// Transform the JSON log string that looks like this:
//
//	{"time":"2024-08-18T12:38:47.271462-07:00","level":"INFO",
//	"source":{"function":"github.com/Cyberax/slog-tidbits/tidbits.TestPrettySink",
//	"file":"/Users/cyberax/bricks/slog-tidbits/tidbits/pretty_sink_test.go","line":16},"msg":"hello, world","key":42}
//
// Into something that looks like this:
func (p *PrettySink) prettyPrint(logStr []byte) error {
	entry := bytes.NewBuffer(nil)
	entry.Grow(len(logStr))

	fields := map[string]any{}

	d := json.NewDecoder(bytes.NewReader(logStr))
	d.UseNumber()
	err := d.Decode(&fields)
	if err != nil {
		return err
	}

	// Make nicer-looking time
	timeValStr := valAsStr(fields, "time") // "time":"2024-08-18T12:38:47.271462-07:00"
	if timeValStr != "" {
		delete(fields, "time")
		entryTime, err := time.Parse(time.RFC3339Nano, timeValStr)
		if err == nil {
			// Format time in a more succinct way
			entry.WriteString(entryTime.Format("15:04:05.000"))
			entry.WriteString(p.separator)
		}
	}

	// Print the level
	levelVal := valAsStr(fields, "level")
	if levelVal != "" {
		delete(fields, "level")
		entry.WriteString(p.formatLevel(levelVal))
		entry.WriteString(p.separator)
	}

	// Print the source
	srcField := fields["source"]
	srcMap, ok := srcField.(map[string]any)
	if ok {
		delete(fields, "source")
		fileName := valAsStr(srcMap, "file")
		fileName = path.Base(fileName)
		line := srcMap["line"]
		entry.WriteString(fmt.Sprintf("%s:%v", fileName, line))
		entry.WriteString(p.separator)
	}

	// Print the message
	msgVal := valAsStr(fields, "msg")
	if msgVal != "" {
		delete(fields, "msg")
		entry.WriteString(msgVal)
		entry.WriteString(p.separator)
	}

	// Print the rest of the fields
	var stack any
	for k, v := range fields {
		if stack == nil && k == StackAttrName {
			stack = v
			continue
		}

		entry.WriteString(k)
		entry.WriteString("=")
		entry.WriteString(fmt.Sprint(v))
		entry.WriteString(p.separator)
	}

	if stack != nil {
		if p.printStack(entry, stack) {
			entry.Truncate(entry.Len() - 1) // Remove the trailing newline
		} else {
			// We failed to do nice stack printing, so just print it as a regular field
			entry.WriteString(StackAttrName)
			entry.WriteString("=")
			entry.WriteString(fmt.Sprint(stack))
			entry.WriteString(p.separator)
			// This is the last element, so we need to remove the trailing separator
			entry.Truncate(entry.Len() - len(p.separator))
		}
	}

	entry.WriteString("\n")
	_, err = p.delegate.Write(entry.Bytes())
	return err
}

func (p *PrettySink) formatLevel(levelVal string) string {
	if !p.colorize {
		return levelVal
	}
	var lvl slog.Level
	err := (&lvl).UnmarshalText([]byte(levelVal))
	if err != nil {
		return levelVal
	}
	if lvl >= slog.LevelError {
		return ColorRed + levelVal + ColorReset
	}
	if lvl >= slog.LevelWarn {
		return ColorYellow + levelVal + ColorReset
	}
	if lvl >= slog.LevelInfo {
		return levelVal
	}
	// Anything less than INFO is debug
	return ColorGray + levelVal + ColorReset
}

func (p *PrettySink) printStack(entry *bytes.Buffer, stack any) bool {
	stackEntry := bytes.NewBuffer(nil)

	stackElements, ok := stack.([]any)
	if !ok {
		return false
	}

	for i, curStackElem := range stackElements {
		elem, ok := curStackElem.(map[string]any)
		if !ok {
			return false
		}
		if i == 0 {
			// First element can be a message
			if panicMsg := valAsStr(elem, "panic_msg"); panicMsg != "" {
				stackEntry.WriteString("\tpanic: " + panicMsg)
				stackEntry.WriteString("\n")
				continue
			}
		}
		fl := valAsStr(elem, "fl")
		fn := valAsStr(elem, "fn")
		if fl == "" || fn == "" {
			return false
		}
		stackEntry.WriteString(fmt.Sprintf("\t%s (%s)\n", fl, fn))
	}

	// Remove the trailing separator
	if entry.Len() > len(p.separator) {
		entry.Truncate(entry.Len() - len(p.separator))
	}
	entry.WriteString("\n")
	entry.WriteString(stackEntry.String())
	return true
}
