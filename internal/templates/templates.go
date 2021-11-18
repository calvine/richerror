package templates

const (
	ErrorConstructorTemplate = `
package {{ .ErrorPkg }}

/* WARNING: This is GENERATED CODE Please do not edit. */

import (
	"github.com/calvine/richerror/errors"

	{{ range getDataItemImportMap .MetaData -}}
		"{{- . -}}"
	{{ end }}
)

// ErrCode{{ .Code }} {{ .Message }}
const ErrCode{{ .Code }} = "{{ .Code }}"

// New{{ .Code }}Error creates a new specific error
func New{{ .Code }}Error({{ range .MetaData }}{{ .Name }} {{ .DataType }}, {{ end }}{{ if .IncludeMap }}fields map[string]interface{}, {{ end }}includeStack bool) errors.RichError {
	msg := "{{ .Message }}"
	err := errors.NewRichError(ErrCode{{ .Code }}, msg)
	{{- if .IncludeMap -}}
		.WithMetaData(fields)
	{{- end -}}
	{{- range .MetaData -}}
	{{- if eq .DataType "error" -}}
		.AddError({{ .Name }})
	{{- else -}}
		.AddMetaData("{{ .Name }}", {{ .Name }})
	{{- end -}}
	{{- end -}}
	{{- if .Tags -}}
		.WithTags([]string{
		{{- range .Tags -}}
			"{{- . -}}",
		{{- end -}}
	})
	{{- end }}
	if includeStack {
		err = err.WithStack(1)
	}
	return err
}

func Is{{ .Code }}Error(err errors.ReadOnlyRichError) bool {
	return err.GetErrorCode() == ErrCode{{ .Code }}
}

`

// TODO: determine if we want the error code in a seperate package.

// 	ErrorCodeTemplate = `
// package {{ .CodePkg }}

// /* WARNING: This is GENERATED CODE Please do not edit. */

// // ErrCode{{ .Code }} {{ .Message }}
// const ErrCode{{ .Code }} = "{{ .Code }}"

// `
)
