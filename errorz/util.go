package errorz

import (
	"runtime"
	"strconv"
	"strings"
)

// Maketrace generates a file name + line number string tracing back to the calling code
// atDepth=0 traces to the call MakeTrace
// atDepth=1 traces to the call before that, useful for using maketrace inside reusable functions
func MakeTrace(atDepth int) string {
	stackDepth := atDepth + 2 // compensate for the 3 function calls that appear beneath the calling context
	pc := make([]uintptr, stackDepth*2)
	n := runtime.Callers(stackDepth, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, ok := frames.Next()
	if !ok {
		return ""
	}

	fileparts := strings.SplitN(frame.File, "alicenet", 2)
	return fileparts[len(fileparts)-1] + ":" + strconv.Itoa(frame.Line)
}
