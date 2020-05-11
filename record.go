package log

import (
	"time"
)

//Record is one log record
type Record struct {
	Time    time.Time
	Caller  Caller
	Level   Level
	Message string
}

//IEncoder ...
type IEncoder interface {
	Encode(l ILogger, r Record) []byte
}
