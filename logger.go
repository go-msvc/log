package log

import (
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"
)

//New creates a standalone logger that writes to the specified writer
func New(w io.Writer) ILogger {
	l := &logger{
		parent:  nil,
		name:    "",
		level:   DebugLevel,
		data:    map[string]interface{}{},
		subs:    map[string]ILogger{},
		writer:  w,
		encoder: DefaultEncoder(),
	}
	return l
} //NewLogger()

//ILogger ...
type ILogger interface {
	Name() string

	//New creates a sub-logger using the same writer
	//and it inherits data from the parent
	New(n string) ILogger

	//Set a name-value
	Set(n string, v interface{})
	//Get name-value
	Get(n string) (interface{}, bool)

	//output functions
	Log(level Level, msg string)
	Trace(msg string)
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)

	//formatted output functions
	Logf(level Level, format string, args ...interface{})
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	//--------------------------------------------------------------------------
	//NOTE: all "With...()" methods updates the current logger and all children
	//--------------------------------------------------------------------------
	//set the level and return the same logger
	//also update all children
	WithLevel(l Level) ILogger

	//With Set(name-value) and delete it from all children
	//then returns the modified logger
	With(n string, v interface{}) ILogger

	//set the encode and return the same logger
	//also update all children
	WithEncoder(e IEncoder) ILogger

	//set the write and return the same logger
	//also update all children
	WithWriter(w io.Writer) ILogger
}

//ValidName is identifier ""
const namePattern = `[a-z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?`

var nameRegex = regexp.MustCompile(`^` + namePattern + `$`)

//ValidName returns true is name is valid
func ValidName(n string) bool {
	return nameRegex.MatchString(n)
}

//logger implements ILogger
type logger struct {
	mutex   sync.Mutex
	parent  ILogger
	name    string
	level   Level
	data    map[string]interface{}
	subs    map[string]ILogger
	writer  io.Writer
	encoder IEncoder
}

func (l *logger) New(n string) ILogger {
	if !ValidName(n) {
		panic("invalid logger name \"" + n + "\"")
	}
	if exists, ok := l.subs[n]; ok {
		return exists
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	sub := &logger{
		parent:  l,
		name:    n,
		level:   l.level,
		data:    map[string]interface{}{},
		subs:    map[string]ILogger{}, //inherits parent's data + own
		writer:  l.writer,             //inherits parent's writer or replace with own
		encoder: l.encoder,
	}
	l.subs[n] = sub
	return sub
} //logger.New()

//Name of this logger
func (l *logger) Name() string {
	if l.parent == nil {
		return l.name
	}
	return l.parent.Name() + "." + l.name
} //logger.Name()

//Set a name=value
func (l *logger) Set(n string, v interface{}) {
	if !ValidName(n) {
		panic(fmt.Sprintf("logger.Set(%s) is invalid name", n))
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.data[n] = v
} //logger.Set()

//Get a data field from self else from parent else nil
func (l *logger) Get(n string) (interface{}, bool) {
	if v, ok := l.data[n]; ok {
		return v, ok
	}
	if l.parent != nil {
		return l.parent.Get(n)
	}
	return nil, false
} //logger.Get()

func (l *logger) log(skip int, level Level, msg string) {
	if l.encoder == nil || l.writer == nil {
		return
	}
	if level >= l.level {
		//gather info for the log record
		record := Record{
			Time:    time.Now(),
			Caller:  GetCaller(skip + 4),
			Level:   level,
			Message: msg,
		}

		//encode and write it
		encodedRecord := l.encoder.Encode(l, record)
		l.writer.Write(encodedRecord)
	}
}

func (l *logger) logf(level Level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log(1, level, msg)
}

func (l *logger) Log(level Level, msg string) { l.log(0, level, msg) }
func (l *logger) Trace(msg string)            { l.log(0, TraceLevel, msg) }
func (l *logger) Debug(msg string)            { l.log(0, DebugLevel, msg) }
func (l *logger) Info(msg string)             { l.log(0, InfoLevel, msg) }
func (l *logger) Warn(msg string)             { l.log(0, WarnLevel, msg) }
func (l *logger) Error(msg string)            { l.log(0, ErrorLevel, msg) }
func (l *logger) Fatal(msg string)            { l.log(0, FatalLevel, msg) }

func (l *logger) Logf(level Level, format string, args ...interface{}) { l.logf(level, format, args...) }
func (l *logger) Tracef(format string, args ...interface{})            { l.logf(TraceLevel, format, args...) }
func (l *logger) Debugf(format string, args ...interface{})            { l.logf(DebugLevel, format, args...) }
func (l *logger) Infof(format string, args ...interface{})             { l.logf(InfoLevel, format, args...) }
func (l *logger) Warnf(format string, args ...interface{})             { l.logf(WarnLevel, format, args...) }
func (l *logger) Errorf(format string, args ...interface{})            { l.logf(ErrorLevel, format, args...) }
func (l *logger) Fatalf(format string, args ...interface{})            { l.logf(FatalLevel, format, args...) }

func (l *logger) WithLevel(level Level) ILogger {
	if level >= _minLevel && level <= _maxLevel {
		l.level = level
		for _, ll := range l.subs {
			ll.WithLevel(level)
		}
	}
	return l
}

func (l *logger) With(n string, v interface{}) ILogger {
	if ValidName(n) {
		if v == nil {
			delete(l.data, n)
		} else {
			l.data[n] = v
		}
		for _, ll := range l.subs {
			ll.With(n, nil) //delete in sub loggers to inherit this value
		}
	}
	return l
} //logger.With()

func (l *logger) WithEncoder(e IEncoder) ILogger {
	if e != nil {
		l.encoder = e
		for _, ll := range l.subs {
			ll.WithEncoder(e)
		}
	}
	return l
}

func (l *logger) WithWriter(w io.Writer) ILogger {
	if w != nil {
		l.writer = w
		for _, ll := range l.subs {
			ll.WithWriter(w)
		}
	}
	return l
}
