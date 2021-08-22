# RichError

This is a package contains my thought on good error practices. It includes an error implementation called richError and two interfaces called RichError and ReadOnlyRichError.

## Goals

The RichError interface and its primary implementation in this module is an attempt to standardize errors to assist in achieveing these goals:

1. Have standardized fields for errors
2. Having consistent error messages and error codes to make reporting and application analysis easier
3. Break out details about error occurances in discrete fields.
4. Nice serializable data.
5. Robust output from error.Error() calls

## Error Definion Schema

``` go
type dataItem struct {
 // Name is the name of the parameter added to the error constructor as well as the label added to the parameter in the errors metadata.
 Name string `json:"name"`
 // DataType is a string that tells the go generator what the type of this field is for the error constructor.
 DataType string `json:"dataType"`
}

type errorData struct {
 // Code is expected to be Pascal Case. Is a preferable unique string code for an error.
 Code string `json:"code"`
 // Tags are a way of grouping errors together so that the can be target for generation in groups.
 Tags []string `json:"tags"`
 // Message is a string added as the message to the error produced.
 Message string `json:"message"`
 // IncludeMap if true adds a map[string]interface{} to the parameters of a constructor so that a genereic map of data can get added to an error constructor parameters list in addition to any specific data defined in MetaData.
 IncludeMap bool `json:"includeMap"`
 // MetaData is an array of dataItem that lists specific data that should be added to the error constructor, and added to the errors metadata map.
 MetaData []dataItem `json:"metaData"`
}
```

## Rich Error Details

`TODO: write this up.`

## Error generator

Currently the code generator is a simple command line app. It can be installed using `go install` and then used from the command line to generate errors for your applications from an error definitions file.

An example of running the code generator:

`go run main.go generate -i "example_errors.json" -o "testapp"`

An example with the code generator installed:

`richerror generate -i "example_errors.json" -o "testapp"`

## Additional language support

Right now there are templates for generating error constructors and codes only for the Go language. In the future I would like to add additional languages. The ideal use case for this would be to maintain a "dictionary" of errors for your application / domain and be able to run the code generator to make nice errors for use in development that will enforce adding the proper data and helping to achieve the goals listed above

When additional languages are supported I envision them as sub commands of the generate command. That way we can have persistent flags to have uniform cli flags to pass down throught to the language specific commands. Then we can have language specific commands to do things like add namespaces for C# or something similar in other languages.
