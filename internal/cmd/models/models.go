package models

type DataItem struct {
	// Name is the name of the parameter added to the error constructor as well as the label added to the parameter in the errors metadata.
	Name string `json:"name"`
	// DataType is a string that tells the go generator what the type of this field is for the error constructor.
	DataType string `json:"dataType"`
	// ImportPath specifies the import path for the data type to be inserted into the error template.
	ImportPath string `json:"importPath"`
}

type ErrorData struct {
	// Code is expected to be Pascal Case. Is a preferable unique string code for an error.
	Code string `json:"code"`
	// Tags are a way of grouping errors together so that the can be target for generation in groups, Also these tags can be used for aggregation in log viewers.
	Tags []string `json:"tags"`
	// Message is a string added as the message to the error produced.
	Message string `json:"message"`
	// IncludeMap if true adds a map[string]interface{} to the parameters of a constructor so that a genereic map of data can get added to an error constructor parameters list in addition to any specific data defined in MetaData.
	IncludeMap bool `json:"includeMap"`
	// MetaData is an array of dataItem that lists specific data that should be added to the error constructor, and added to the errors metadata map.
	MetaData []DataItem `json:"metaData"`
}

type GeneratorData struct {
	ErrorPkg string
	ErrorData
}
