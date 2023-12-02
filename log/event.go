package log

import (
	"fmt"
	"time"
)

type Event struct {
	Level   Level
	Message string
	Time    time.Time
}

func newEvent(level Level, format string, args ...any) *Event {
	event := &Event{
		Level:   level,
		Time:    time.Now(),
		Message: fmt.Sprintf(format, args...),
	}

	return event
}
