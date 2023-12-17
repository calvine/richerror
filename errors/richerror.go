package errors

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type RichErrorOutputFormat int
type CustomOutputFunc func(e ReadOnlyRichError) string

var (
	// customOutputFunction is a global function for a custom output format for rich errors in a text format.
	// it is set by calling SetGlobalCustomOutputFunction
	// the custom output function can also be set on the error level by calling the SetCustomOutputFunction function
	customOutputFunction CustomOutputFunc
	// errorOutputFormat is a global setting for the default output format of a rich error in text format.
	// it is set by calling SetGlobalErrorOutputFormat
	// the output format can also be set on the specific error level via the SetOutputFormat function
	errorOutputFormat RichErrorOutputFormat = FullOutputFormatted
)

const (
	NotSpecified RichErrorOutputFormat = iota
	CustomOutput
	DetailedOutput
	FullOutputFormatted
	FullOutputInline
	ShortDetailedOutput
	ShortOutput
)

type ReadOnlyRichError interface {
	GetErrorCode() string
	GetErrorMessage() string
	GetStack() []callStackEntry
	GetSource() string
	GetFunction() string
	GetLineNumber() string
	GetTags() []string
	GetMetaData() map[string]interface{}
	GetMetaDataItem(key string) (interface{}, bool)
	GetErrors() []error
	HasStack() bool
	ToString(format RichErrorOutputFormat) string
	ToCustomString() string

	error
}

type RichError interface {
	WithStack(stackOffset int) RichError
	WithMetaData(metaData map[string]interface{}) RichError
	WithErrors(errs []error) RichError
	WithTags(tags []string) RichError
	AddSource(source string) RichError
	AddFunction(function string) RichError
	AddLineNumber(lineNumber string) RichError
	AddMetaData(key string, value interface{}) RichError
	AddError(err error) RichError
	AddTag(tag string) RichError

	SetCustomOutputFunction(cof CustomOutputFunc) RichError
	SetOutputFormat(outputFormat RichErrorOutputFormat) RichError

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
	ErrCode              string                 `json:"code"`
	Message              string                 `json:"message"`
	Source               string                 `json:"source,omitempty"`
	Function             string                 `json:"function,omitempty"`
	Line                 string                 `json:"line,omitempty"`
	OccurredAt           time.Time              `json:"occurredAt"`
	Tags                 []string               `json:"tags"`
	Stack                []callStackEntry       `json:"stack,omitempty"`
	InnerErrors          []error                `json:"innerErrors"`
	MetaData             map[string]interface{} `json:"metaData"`
	outputFormat         RichErrorOutputFormat  `json:"-"`
	customOutputFunction CustomOutputFunc       `json:"-"`
}

func SetGlobalCustomOutputFunction(cof CustomOutputFunc) {
	customOutputFunction = cof
}

func SetGlobalErrorOutputFormat(format RichErrorOutputFormat) {
	errorOutputFormat = format
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

			functionName := nextFrame.Function
			if len(functionName) > 0 {
				functionNameLastIndex := strings.LastIndex(functionName, ".")
				functionName = functionName[functionNameLastIndex+1:]
			}
			e.Source = source
			e.Function = functionName
			e.Line = strconv.Itoa(nextFrame.Line)
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
	e.InnerErrors = append(e.InnerErrors, errs...)
	return e
}

func (e richError) WithTags(tags []string) RichError {
	e.Tags = tags
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
	e.Line = lineNumber
	return e
}

func (e richError) AddMetaData(key string, value interface{}) RichError {
	if e.MetaData == nil {
		e.MetaData = make(map[string]interface{})
	}
	e.MetaData[key] = value
	return e
}

func (e richError) AddError(err error) RichError {
	if err != nil {
		e.InnerErrors = append(e.InnerErrors, err)
	}
	return e
}

func (e richError) AddTag(tag string) RichError {
	e.Tags = append(e.Tags, tag)
	return e
}

func (e richError) SetCustomOutputFunction(cof CustomOutputFunc) RichError {
	e.customOutputFunction = cof
	return e
}

func (e richError) SetOutputFormat(outputFormat RichErrorOutputFormat) RichError {
	e.outputFormat = outputFormat
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
	return e.Line
}

func (e richError) GetMetaData() map[string]interface{} {
	return e.MetaData
}

func (e richError) GetTags() []string {
	return e.Tags
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

func (e richError) getCustomOutputFunction() CustomOutputFunc {
	if e.customOutputFunction != nil {
		return e.customOutputFunction
	}
	return customOutputFunction
}

func (e richError) getErrorOutputFormat() RichErrorOutputFormat {
	if e.outputFormat != NotSpecified {
		return e.outputFormat
	}
	return errorOutputFormat
}

func (e richError) HasStack() bool {
	return len(e.Stack) > 0
}

func (e richError) ToString(format RichErrorOutputFormat) string {
	switch format {
	case CustomOutput:
		return e.ToCustomString()
	case DetailedOutput:
		return e.detailedOutputString("\n", "\t")
	case FullOutputFormatted:
		return e.fullOutputString("\n", "\t")
	case FullOutputInline:
		return e.fullOutputString(" --- ", "")
	case ShortDetailedOutput:
		return e.shortDetailedOutputString(" - ")
	default: // ShortOutput is default?
		return e.shortOutputString(" - ")
	}
}

func (e richError) ToCustomString() string {
	cof := e.getCustomOutputFunction()
	if cof == nil {
		panic("CustomOutput mode is selected and no custom output function set for the error or globally")
	}
	return cof(e)
}

func (e richError) Error() string {
	eof := e.getErrorOutputFormat()
	return e.ToString(eof)
}

func (e richError) shortOutputString(separator string) string {
	return fmt.Sprintf("%s%s%s%s%s", e.OccurredAt.String(), separator, e.ErrCode, separator, e.Message)
}

func (e richError) shortDetailedOutputString(separator string) string {
	return fmt.Sprintf("%s%s%s%s%s%s%s:%s", e.OccurredAt.String(), separator, e.ErrCode, separator, e.Message, separator, e.Source, e.Line)
}

func (e richError) detailedOutputString(partSeparator, indentString string) string {
	var messageBuffer bytes.Buffer
	timeStampMsg := fmt.Sprintf("ERROR - %s", e.OccurredAt.String())
	messageBuffer.WriteString(timeStampMsg)
	if e.Source != "" {
		sourceSection := fmt.Sprintf("%sSOURCE: %s:%s", partSeparator, e.Source, e.Line)
		messageBuffer.WriteString(sourceSection)
	}
	if e.ErrCode != "" {
		errCodeSection := fmt.Sprintf("%sERRCODE: %s", partSeparator, e.ErrCode)
		messageBuffer.WriteString(errCodeSection)
	}
	if e.Message != "" {
		messageSection := fmt.Sprintf("%sMESSAGE: %s", partSeparator, e.Message)
		messageBuffer.WriteString(messageSection)
	}
	if len(e.MetaData) > 0 {
		messageBuffer.WriteString("METADATA:")
		for key, value := range e.MetaData {
			metaDataMsg := fmt.Sprintf("%s%s%s: %v", partSeparator, indentString, key, value)
			messageBuffer.WriteString(metaDataMsg)
		}
	}
	return messageBuffer.String()
}

func (e richError) fullOutputString(partSeparator, indentString string) string {
	var messageBuffer bytes.Buffer
	timeStampMsg := fmt.Sprintf("TIMESTAMP: %s", e.OccurredAt.String())
	messageBuffer.WriteString(timeStampMsg)
	if e.Source != "" {
		sourceSection := fmt.Sprintf("%sSOURCE: %s", partSeparator, e.Source)
		messageBuffer.WriteString(sourceSection)
	}
	if e.Function != "" {
		functionSection := fmt.Sprintf("%sFUNCTION: %s", partSeparator, e.Function)
		messageBuffer.WriteString(functionSection)
	}
	if e.Line != "" {
		LineNumberSection := fmt.Sprintf("%sLINE_NUM: %s", partSeparator, e.Line)
		messageBuffer.WriteString(LineNumberSection)
	}
	if e.ErrCode != "" {
		errCodeSection := fmt.Sprintf("%sERRCODE: %s", partSeparator, e.ErrCode)
		messageBuffer.WriteString(errCodeSection)
	}
	if e.Message != "" {
		messageSection := fmt.Sprintf("%sMESSAGE: %s", partSeparator, e.Message)
		messageBuffer.WriteString(messageSection)
	}
	if len(e.Stack) > 0 {
		stackBuffer := bytes.Buffer{}
		firstLine := fmt.Sprintf("%sSTACK: ", partSeparator)
		stackBuffer.WriteString(firstLine)
		for _, frame := range e.Stack {
			stackFrame := fmt.Sprintf("%s%s%s", strings.Repeat(indentString, frame.Depth), frame.String(), partSeparator)
			stackBuffer.WriteString(stackFrame)
		}
		messageBuffer.WriteString(stackBuffer.String())
	}
	if len(e.InnerErrors) > 0 {
		messageBuffer.WriteString("INNER ERRORS:")
		for i, err := range e.InnerErrors {
			if err != nil { // errors here should never be nil but just incase?
				messageBuffer.WriteString(getInnerErrorString(err, partSeparator, indentString, i))
			}
		}
		messageBuffer.WriteString(partSeparator)
	}
	if len(e.MetaData) > 0 {
		messageBuffer.WriteString("METADATA:")
		for key, value := range e.MetaData {
			metaDataMsg := fmt.Sprintf("%s%s%s: %v", partSeparator, indentString, key, value)
			messageBuffer.WriteString(metaDataMsg)
		}
	}
	return messageBuffer.String()
}

func getInnerErrorString(err error, partSeparator string, indentString string, index int) string {
	var innerErrString string
	if richError, ok := err.(ReadOnlyRichError); ok {
		innerErrString = fmt.Sprintf("%s%sERROR #%d: %s", partSeparator, strings.Repeat(indentString, index+1), index+1, richError.ToString(ShortDetailedOutput))
	} else {
		innerErrString = fmt.Sprintf("%s%sERROR #%d: %s", partSeparator, strings.Repeat(indentString, index+1), index+1, err.Error())
	}
	return innerErrString
}
