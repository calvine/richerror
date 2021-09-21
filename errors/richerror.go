package errors

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	indentString  = "\t" //""
	partSeperator = "\n" //" - "
)

type ReadOnlyRichError interface {
	GetErrorCode() string
	GetErrorMessage() string
	GetStack() []callStackEntry
	GetSource() string
	GetFunction() string
	GetLineNumber() string
	GetMetaData() map[string]interface{}
	GetMetaDataItem(key string) (interface{}, bool)
	GetErrors() []error

	error
}

type RichError interface {
	WithStack(stackOffset int) RichError
	WithMetaData(metaData map[string]interface{}) RichError
	WithErrors(errs []error) RichError
	AddSource(source string) RichError
	AddFunction(function string) RichError
	AddLineNumber(lineNumber string) RichError
	AddMetaData(key string, value interface{}) RichError
	AddError(err error) RichError

	ReadOnlyRichError
}

type callStackEntry struct {
	Depth    int     `json:"depth"`
	Entry    uintptr `json:"entry"`
	File     string  `json:"file"`
	Function string  `json:"function"`
	Line     int     `json:"line"`
	PC       uintptr `json:"pc"`
}

func (cse *callStackEntry) String() string {
	return fmt.Sprintf("L:%d %v - %s:%d - %s", cse.Depth, cse.Entry, cse.File, cse.Line, cse.Function)
}

type richError struct {
	ErrCode     string                 `json:"code"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source,omitempty"`
	Function    string                 `json:"function,omitempty"`
	LineNumber  string                 `json:"lineNumber,omitempty"`
	Stack       []callStackEntry       `json:"stack,omitempty"`
	OccurredAt  time.Time              `json:"occurredAt"`
	InnerErrors []error                `json:"innerErrors"`
	MetaData    map[string]interface{} `json:"metaData"`
}

func NewRichError(errCode, message string) RichError {
	occurredAt := time.Now().UTC()
	err := richError{
		ErrCode:    errCode,
		Message:    message,
		OccurredAt: occurredAt,
	}
	return err

}

func NewRichErrorWithStack(errCode, message string, stackOffset int) RichError {
	err := NewRichError(errCode, message).WithStack(stackOffset)
	return err
}

func (e richError) WithStack(stackOffset int) RichError {
	baseStackOffset := 2
	// Here we initialize the slice to 10 because the runtime.Callers
	// function will not grow the slice as needed.
	var callerData []uintptr = make([]uintptr, 10)
	// Here we use 2 to remove the runtime.Callers call
	// and the call to the RichError.WithStack call.
	// This should leave only the relevant stack pieces
	numFrames := runtime.Callers(baseStackOffset+stackOffset, callerData)
	data := runtime.CallersFrames(callerData)
	for i := 0; i < numFrames; i++ {
		nextFrame, _ := data.Next()
		if i == 0 {
			source := nextFrame.File
			// if len(source) > 0 {
			// 	sourceLastIndex := strings.LastIndex(nextFrame.File, string(os.PathSeparator))
			// 	source = source[sourceLastIndex+1:]
			// }

			functionName := nextFrame.Function
			if len(functionName) > 0 {
				functionNameLastIndex := strings.LastIndex(functionName, ".")
				functionName = functionName[functionNameLastIndex+1:]
			}
			e.Source = source
			e.Function = functionName
			e.LineNumber = strconv.Itoa(nextFrame.Line)
		}
		callStackEntry := callStackEntry{
			Depth:    i,
			Entry:    nextFrame.Entry,
			File:     nextFrame.File,
			Function: nextFrame.Function,
			Line:     nextFrame.Line,
			PC:       nextFrame.PC,
		}
		e.Stack = append(e.Stack, callStackEntry)
	}

	return e
}

func (e richError) WithMetaData(metaData map[string]interface{}) RichError {
	e.MetaData = metaData
	return e
}

func (e richError) WithErrors(errs []error) RichError {
	if errs != nil {
		if e.InnerErrors == nil {
			e.InnerErrors = make([]error, len(errs))
		}
		e.InnerErrors = append(e.InnerErrors, errs...)
	}
	return e
}

func (e richError) AddSource(source string) RichError {
	e.Source = source
	return e
}

func (e richError) AddFunction(function string) RichError {
	e.Function = function
	return e
}

func (e richError) AddLineNumber(lineNumber string) RichError {
	e.LineNumber = lineNumber
	return e
}

func (e richError) AddMetaData(key string, value interface{}) RichError {
	if e.MetaData == nil {
		e.MetaData = make(map[string]interface{}, 1)
	}
	e.MetaData[key] = value
	return e
}

func (e richError) AddError(err error) RichError {
	if err != nil {
		if e.InnerErrors == nil {
			e.InnerErrors = make([]error, 1)
		}
		e.InnerErrors = append(e.InnerErrors, err)
	}
	return e
}

func (e richError) GetErrorCode() string {
	return e.ErrCode
}

func (e richError) GetErrorMessage() string {
	return e.Message
}

func (e richError) GetStack() []callStackEntry {
	return e.Stack
}

func (e richError) GetSource() string {
	return e.Source
}

func (e richError) GetFunction() string {
	return e.Function
}

func (e richError) GetLineNumber() string {
	return e.LineNumber
}

func (e richError) GetMetaData() map[string]interface{} {
	return e.MetaData
}

func (e richError) GetMetaDataItem(key string) (interface{}, bool) {
	if e.MetaData == nil {
		return nil, false
	}
	val, ok := e.MetaData[key]
	return val, ok
}

func (e richError) GetErrors() []error {
	return e.InnerErrors
}

func (e richError) Error() string {
	var messageBuffer bytes.Buffer
	timeStampMsg := fmt.Sprintf("TIMESTAMP: %s", e.OccurredAt.String())
	messageBuffer.WriteString(timeStampMsg)
	if e.Source != "" {
		sourceSection := fmt.Sprintf("%sSOURCE: %s", partSeperator, e.Source)
		messageBuffer.WriteString(sourceSection)
	}
	if e.Function != "" {
		functionSection := fmt.Sprintf("%sFUNCTION: %s", partSeperator, e.Function)
		messageBuffer.WriteString(functionSection)
	}
	if e.LineNumber != "" {
		LineNumberSection := fmt.Sprintf("%sLINE_NUM: %s", partSeperator, e.LineNumber)
		messageBuffer.WriteString(LineNumberSection)
	}
	if e.ErrCode != "" {
		errCodeSection := fmt.Sprintf("%sERRCODE: %s", partSeperator, e.ErrCode)
		messageBuffer.WriteString(errCodeSection)
	}
	if e.Message != "" {
		messageSection := fmt.Sprintf("%sMESSAGE: %s", partSeperator, e.Message)
		messageBuffer.WriteString(messageSection)
	}
	if len(e.Stack) > 0 {
		stackBuffer := bytes.Buffer{}
		firstLine := fmt.Sprintf("%sSTACK: ", partSeperator)
		stackBuffer.WriteString(firstLine)
		for _, frame := range e.Stack {
			stackFrame := fmt.Sprintf("%s%s%s", strings.Repeat(indentString, frame.Depth), frame.String(), partSeperator)
			stackBuffer.WriteString(stackFrame)
		}
		messageBuffer.WriteString(stackBuffer.String())
	}
	if len(e.InnerErrors) > 0 {
		messageBuffer.WriteString("INNER ERRORS:")
		for i, err := range e.InnerErrors {
			innerErrMessage := fmt.Sprintf("%s%sERROR #%d: %s", partSeperator, strings.Repeat(indentString, i+1), i+1, err.Error())
			messageBuffer.WriteString(innerErrMessage)
		}
		messageBuffer.WriteString(partSeperator)
	}
	if len(e.MetaData) > 0 {
		messageBuffer.WriteString("METADATA:")
		for key, value := range e.MetaData {
			metaDataMsg := fmt.Sprintf("%s%s%s: %v", partSeperator, indentString, key, value)
			messageBuffer.WriteString(metaDataMsg)
		}
	}
	return messageBuffer.String()
}
