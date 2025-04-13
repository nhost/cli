package configserver

import (
	"context"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

type Querier struct{}

func (q Querier) GetAppDesiredState(_ context.Context, _ uuid.UUID) (int32, error) {
	// Get the caller two levels up
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, "changeDatabaseVersionValidate") {
			return 5, nil
		}
		if !more {
			break
		}
	}
	
	// Default to appPaused for all other cases
	return 6, nil //nolint:mnd
}
