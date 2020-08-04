package commands

import "fmt"

// UsageError represents an error that occured due to a command receiving
// invlid arguments or types
type UsageError struct {
	Param    string
	Message  string
	Footer   string
	Provided interface{}
}

func (ue UsageError) Error() string {
	return fmt.Sprintf("<%s>: %s", ue.Param, ue.Message)
}

// SystemError represents an error that has occured with the application
// when running a command and may not (proply is not) the fault of the user
type SystemError struct {
	error
	Message string
	Stack   []byte
}

func (se SystemError) Unwrap() error {
	return se.error
}

func (se SystemError) Error() string {
	return se.Message
}

// Warning represents an warning that a command could not be completed
// due to the commands logic
type Warning struct {
	Message string
}

func (ge Warning) Error() string {
	return ge.Message
}

// PermissionError represents an error that occurs when a user tries to use an
// action or command they are not allowed to use
type PermissionError struct {
	Message string
}

func (pe PermissionError) Error() string {
	return pe.Message
}
