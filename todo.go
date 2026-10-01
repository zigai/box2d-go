package box2d

import (
	"fmt"
	"runtime"
	"strings"
)

type unported struct {
	fn string
}

func (u unported) Unported() string { return u.fn }

func (u unported) Error() string { return fmt.Sprintf("box2d: %s is not ported yet", u.fn) }

func todo() {
	name := "unknown"
	var pcs [1]uintptr

	if runtime.Callers(2, pcs[:]) == 1 {
		frame, _ := runtime.CallersFrames(pcs[:]).Next()
		if frame.Function != "" {
			name = strings.TrimPrefix(frame.Function, "github.com/zigai/")
		}
	}
	panic(unported{fn: name})
}
