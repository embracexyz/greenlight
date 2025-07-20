package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

const (
	INFO Level = iota
	WARNING
	ERROR
	FATAL
	OFF
)

type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

func (l Level) String() string {
	switch l {
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return ""
	}
}

func (l *Logger) Print(level Level, message string, properties map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().Format(time.DateTime),
		Message:    message,
		Properties: properties,
	}
	if level >= ERROR {
		aux.Trace = string(debug.Stack())
	}

	var line []byte
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(level.String() + ": unable to marshal log message: " + err.Error())
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out.Write(append(line, '\n'))
}

func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.Print(INFO, message, properties)
}

func (l *Logger) PrintWarning(message string, properties map[string]string) {
	l.Print(WARNING, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.Print(ERROR, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.Print(FATAL, err.Error(), properties)
	os.Exit(1)
}

func (l *Logger) Write(message []byte) (n int, err error) {
	return l.Print(ERROR, string(message), nil)
}
