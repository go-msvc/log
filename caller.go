package log

import (
	"path"
	"runtime"
	"strings"
)

//Caller refers to code that wrote the log record
type Caller struct {
	Package  string
	Function string
	File     string
	Line     int
}

//GetCaller skipping N levels in call stack
func GetCaller(skip int) Caller {
	caller := Caller{
		Package:  "N/A",
		Function: "N/A",
		File:     "N/A",
		Line:     -1,
	}

	{
		//get call stack details
		//note: pc array size determines max depth call stack retrieved
		pc := make([]uintptr, 10)
		//n is nr of items retrieved
		n := runtime.Callers(0, pc)
		//fmt.Printf("Got n=%d frames:\n", n)
		//to get details of each pc in the stack, use:
		//		frames := runtime.CallersFrames(pc)
		//it will output the following:
		//		Frame[0] = runtime.Callers
		//		Frame[1] = the line above where we called that runtime.Callers()
		//		Frame[2] = logger._record() = the caller of this func
		//		Frame[3] = logger.log/logf = called of record
		//		Frame[4] = logger.Debug()/Info()/...
		//		Frame[5] = the one we want!!!
		//we are only interested in [5], so cannot retrieve if stack is shallower

		//print the whole stack for debugging this func:
		// if true {
		// 	frames := runtime.CallersFrames(pc)
		// 	i := 0
		// 	for {
		// 		frame, more := frames.Next()
		// 		fmt.Printf("  Frame[%d]: %+v Mode:%v Func=%v File=%s(%d)\n", i, frame, more, frame.Function, frame.File, frame.Line)
		// 		if !more {
		// 			break
		// 		}
		// 		i++
		// 	}
		// }

		if n >= skip {
			pc = pc[skip : skip+1]
			frames := runtime.CallersFrames(pc)
			frame, _ := frames.Next()

			//function is "<package>.<func>" and <package> is path notation that may contain more '.'
			//get basename of package then split on '.'
			d := path.Dir(frame.Function)
			if d == "." {
				d = ""
			}
			b := path.Base(frame.Function)
			p := strings.SplitN(b, ".", 2)
			if len(p) == 2 {
				caller.Package = d + "/" + p[0]
				caller.Function = p[1]
			} else {
				caller.Package = "?"
				caller.Function = frame.Function
			}

			//so we need to split on the last '.'
			// lastDotIndex := strings.LastIndex(frame.Function, ".")
			// if lastDotIndex >= 0 {
			// 	caller.Package = frame.Function[0:lastDotIndex]
			// 	caller.Function = frame.Function[lastDotIndex+1:]
			// } else {
			// 	caller.Package = ""
			// 	caller.Function = frame.Function
			// }
			caller.File = frame.File
			caller.Line = frame.Line
		} //if stack is deep enough
	} //scope
	return caller
} //GetCaller()
