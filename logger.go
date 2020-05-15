package log

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

//ForThisPackage returns a logger for your package
func ForThisPackage() ILogger {
	c := GetCaller(3)
	fmt.Fprintf(os.Stderr, "ForThisPackage(%s)\n", c.Package)
	return Logger(c.Package)
}

//Top returns the top logger
//set level/writer on this to change all loggers
func Top() ILogger {
	fmt.Fprintf(os.Stderr, "Top()\n")
	return top
}

//Logger returns a named logger
//all loggers are added in a path hierarchy to the top logger
//The name may be any path, but preferably the source path
//e.g. "github.com/go-msvc/logger" which you can get by
//calling logger.ForThisPackage()
func Logger(name string) ILogger {
	fmt.Fprintf(os.Stderr, "Logger(%s)\n", name)
	//if name has paths, create parents first
	names := strings.SplitN(name, "/", -1)
	log := top
	for len(names) > 0 {
		if len(names[0]) > 0 {
			log = log.Logger(names[0])
		}
		names = names[1:]
	}
	if log == top {
		panic("missing logger name")
	}
	return log
} //New()

//ILogger ...
type ILogger interface {
	Name() string

	//Logger creates a sub-logger using the same writer
	//and it inherits data from the parent
	//Parameter n must be a simple name - no path '/' characters
	Logger(n string) ILogger

	//Set a name-value and remove it from all children
	//Set with v=nil also deletes a value for this and all children
	//With is same as Set but return the logger to chain operations
	Set(n string, v interface{})
	With(n string, v interface{}) ILogger
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
	//NOTE: all "Set...()" and "With...()" methods updates the current logger and all children
	// Loggers are not copied as they all exist in the tree
	// With...() is only offered to chain operations, but they do the same as Set...()
	//--------------------------------------------------------------------------
	//set the level and return the same logger
	//also update all children
	SetLevel(l Level)
	WithLevel(l Level) ILogger

	//set the encode and return the same logger
	//also update all children
	SetEncoder(e IEncoder)
	WithEncoder(e IEncoder) ILogger

	//set the write and return the same logger
	//also update all children
	SetWriter(w io.Writer)
	WithWriter(w io.Writer) ILogger
}

//ValidName is a domain name identifier ""
const namePattern = `[a-z]([a-zA-Z0-9\._-]*[a-zA-Z0-9])?`

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

func (l *logger) Logger(n string) ILogger {
	fmt.Fprintf(os.Stderr, "logger(%s).Logger(%s)\n", l.Name(), n)
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
	fmt.Fprintf(os.Stderr, "Created logger(%s)\n", sub.Name())
	return sub
} //logger.Logger()

//Name of this logger
func (l *logger) Name() string {
	if l.parent == nil {
		return l.name
	}
	return l.parent.Name() + "/" + l.name
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
		cleanMessage := strings.Map(func(r rune) rune {
			if unicode.IsGraphic(r) {
				return r
			}
			return -1
		}, msg)
		record := Record{
			Time:    time.Now(),
			Caller:  GetCaller(skip + 4),
			Level:   level,
			Message: cleanMessage,
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

func (l *logger) SetLevel(level Level) {
	if level >= _minLevel && level <= _maxLevel {
		l.level = level
		for _, ll := range l.subs {
			ll.WithLevel(level)
		}
	}
} //logger.SetLevel()

func (l *logger) WithLevel(level Level) ILogger {
	l.SetLevel(level)
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

func (l *logger) SetEncoder(e IEncoder) {
	if e != nil {
		l.encoder = e
		for _, ll := range l.subs {
			ll.WithEncoder(e)
		}
	}
}

func (l *logger) WithEncoder(e IEncoder) ILogger {
	l.SetEncoder(e)
	return l
}

func (l *logger) SetWriter(w io.Writer) {
	if w != nil {
		l.writer = w
		for _, ll := range l.subs {
			ll.WithWriter(w)
		}
	}
}

func (l *logger) WithWriter(w io.Writer) ILogger {
	l.SetWriter(w)
	return l
}

//top is the parent of all loggers, allowing any program to discover
//loggers created in various packages using the same logger library
//if you modify settings in a parent (like top) then itself and all
//children are updated, so you can switch on all logging by updating
//top
var (
	top ILogger
	log ILogger
)

func init() {
	//create default top logger
	top = &logger{
		parent:  top,
		name:    "",
		level:   DebugLevel,
		data:    map[string]interface{}{},
		subs:    map[string]ILogger{},
		writer:  os.Stderr,
		encoder: DefaultEncoder(),
	}
	log = ForThisPackage()
}
