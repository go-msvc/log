package log

import (
	"fmt"
	"os"
)

//ILogger ...
type ILogger interface {
	Name() string
	Named(n string) ILogger
	With(n string, v interface{}) ILogger
	Get(n string) interface{}
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	SyncLogs()
}

//Logger ...
type Logger struct {
	parent *Logger
	name   string
	data   map[string]interface{}
}

//NewLogger ...
func NewLogger() ILogger {
	return &Logger{
		parent: nil,
		name:   "",
		data:   map[string]interface{}{},
	}
} //NewLogger()

//Named ...
func (l *Logger) Named(n string) ILogger {
	return &Logger{
		parent: l,
	}
} //Logger.Named()

//Name is full name based on parents with dot-notation
func (l *Logger) Name() string {
	if l.parent == nil {
		return l.name
	}
	return l.parent.Name() + "." + l.name
} //Logger.Name()

//With ...
func (l *Logger) With(n string, v interface{}) ILogger {
	l.data[n] = v
	return l
} //Logger.With()

//Get a data field from self else from parent else nil
func (l *Logger) Get(n string) interface{} {
	if v, ok := l.data[n]; ok {
		return v
	}
	if l.parent != nil {
		return l.parent.Get(n)
	}
	return nil
} //Logger.Get()

//Logf ...
func (l *Logger) Logf(level string, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "(%s)(%+v) %5.5s %s\n", l.Name(), l.data, level, line)
}

//Debugf ...
func (l *Logger) Debugf(format string, args ...interface{}) {
	Logf("debug", format, args...)
}

//Infof ...
func (l *Logger) Infof(format string, args ...interface{}) {
	Logf("info", format, args...)
}

//Warnf ...
func (l *Logger) Warnf(format string, args ...interface{}) {
	Logf("warn", format, args...)
}

//Errorf ...
func (l *Logger) Errorf(format string, args ...interface{}) {
	Logf("error", format, args...)
}

//Fatalf ...
func (l *Logger) Fatalf(format string, args ...interface{}) {
	Logf("fatal", format, args...)
}

//SyncLogs ...
func (l *Logger) SyncLogs() {

}
