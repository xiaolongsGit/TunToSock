package log

import (
	"fmt"
	"strings"
)

type Level int

const (
	SilentLevel Level = iota
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case SilentLevel:
		return "silent"
	default:
		return fmt.Sprintf("not a valid level %d", level)
	}
}

func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "silent":
		return SilentLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	default:
		return Level(0), fmt.Errorf("not a valid logrus Level: %q", lvl)
	}
}
