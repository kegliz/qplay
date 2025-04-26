package dag

import "fmt"

// Public error helpers so callers can assert specific failures.
var (
	ErrBadQubit = fmt.Errorf("builder: qubit index out of range")
	ErrBadClbit = fmt.Errorf("builder: classical bit index out of range")
	ErrSpan     = fmt.Errorf("builder: gate spans invalid qubit range")
	ErrBuild    = fmt.Errorf("builder: cannot build due to previous error")
)
var (
	ErrValidated = fmt.Errorf("dag: already validated, no further mutation")
)
