package errors

import "fmt"

type ErrorType string

const (
	SyntaxError           ErrorType = "SyntaxError"
	StackOverflowError    ErrorType = "StackOverflowError"
	FunctionNotFoundError ErrorType = "FunctionNotFoundError"
	FunctionCallError     ErrorType = "FunctionCallError"
)

type Error struct {
	Type    ErrorType `json:"type"`
	Step    string    `json:"step"`
	Message string    `json:"message"`
	Detail  any       `json:"detail"`
	Trace   []string  `json:"trace"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s (at step: %s), detail: %v", e.Type, e.Message, e.Step, e.Detail)
}

func New(typ ErrorType, step string, message string, detail any, trace []string) *Error {
	return &Error{
		typ,
		step,
		message,
		detail,
		trace,
	}
}
