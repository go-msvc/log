package log

import (
	"fmt"
	"os"
)

//Logf ...
func Logf(level string, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%5.5s %s\n", level, line)
}

//Debugf ...
func Debugf(format string, args ...interface{}) {
	Logf("debug", format, args...)
}

//Infof ...
func Infof(format string, args ...interface{}) {
	Logf("info", format, args...)
}

//Warnf ...
func Warnf(format string, args ...interface{}) {
	Logf("warn", format, args...)
}

//Errorf ...
func Errorf(format string, args ...interface{}) {
	Logf("error", format, args...)
}

//Fatalf ...
func Fatalf(format string, args ...interface{}) {
	Logf("fatal", format, args...)
}
