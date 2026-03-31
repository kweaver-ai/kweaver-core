// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

//go:generate mockgen -source=error.go -destination=../mocks/error.go -package=mocks
import (
	"fmt"
	"net/http"
	"regexp"

	jsoniter "github.com/json-iterator/go"
)

// EOut is a error format for output
type EOut struct {
	isFromNew   bool
	HTTPCode    int
	Code        string      `json:"code"`
	Description string      `json:"description"`
	Solution    string      `json:"solution"`
	Link        string      `json:"link"`
	Detail      interface{} `json:"detail"`
	Err         error
	Fields      map[string]interface{}
}

var (
	ecodeRule            = regexp.MustCompile(`^[A-Z][A-Za-z0-9.]{0,34}$`)
	ErrCodeOpenAPIParser = newEOut("OpenAPIParser", http.StatusBadRequest)
)

func newEOut(code string, httpCode int) *EOut {
	eout := &EOut{}
	eout.Code = mustECode(code)
	eout.HTTPCode = httpCode
	eout.Solution = fmt.Sprintf("Error.%s.Solution", code)
	eout.Description = fmt.Sprintf("Error.%s.Description", code)
	return eout
}

func mustECode(str string) string {
	if !ecodeRule.MatchString(str) {
		panic(fmt.Sprintf("%s is incorrect, ecode should match %s", str, ecodeRule.String()))
	}
	return str
}

// Error output error string
func (e *EOut) Error() string {
	return fmt.Sprintf(e.Description, e.Fields)
}

// New generate a new EOut
func (e *EOut) New(err error) *EOut {
	ne := *e
	ne.isFromNew = true
	ne.Err = err
	ne.Fields = map[string]interface{}{}
	return &ne
}

func (e *EOut) AddField(k string, v interface{}) *EOut {
	if !e.isFromNew {
		panic("eout must init use New()")
	}
	e.Fields[k] = v
	return e
}

// Unwrap unwrap error
func (e *EOut) Unwrap() error {
	return e.Err
}

// CommonError common error
type CommonError struct {
	ErrorCode    string      `json:"ErrorCode"`
	Description  string      `json:"Description"`
	ErrorDetails interface{} `json:"ErrorDetails"`
	ErrorLink    string      `json:"ErrorLink"`
	Solution     string      `json:"Solution"`
}

// Error output error string
func (err *CommonError) Error() string {
	return err.Description
}

type NumCodeError struct {
	Code    int64       `json:"code"`
	Detail  interface{} `json:"detail"`
	Message string      `json:"message"`
	Cause   string      `json:"cause"`
}

func (err *NumCodeError) Error() string {
	return err.Message
}

func GetErrorCause(err error) (msg string) {
	eout, ok := err.(*EOut)
	if ok {
		err = eout.Unwrap()
	}
	nErr, ok := err.(*NumCodeError)
	if ok {
		msg, _ = jsoniter.MarshalToString(nErr)
		return
	}
	cErr, ok := err.(*CommonError)
	if ok {
		msg, _ = jsoniter.MarshalToString(cErr)
		return
	}

	msg = err.Error()
	return
}

type HTTPError struct {
	HTTPCode  int
	Language  string
	BaseError BaseError
}

// BaseError http error
type BaseError struct {
	Code                    string         `json:"code"`
	Description             string         `json:"description"` // Error description
	Solution                string         `json:"solution"`    // Solution
	ErrorLink               string         `json:"link"`        // Error link
	ErrorDetails            interface{}    `json:"details"`     // Detail content
	DescriptionTemplateData map[string]any `json:"-"`           // Description parameters
	SolutionTemplateData    map[string]any `json:"-"`           // Solution parameters
}

// Error return error message
func (err *HTTPError) Error() string {
	errBys, _ := jsoniter.Marshal(err.BaseError)
	return string(errBys)
}
