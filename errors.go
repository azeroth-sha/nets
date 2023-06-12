package nets

import "errors"

var (
	ErrRunning  = errors.New(`service is running`)
	ErrShutdown = errors.New(`service has been shut down`)
)