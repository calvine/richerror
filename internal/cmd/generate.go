/*
Copyright © 2021 Calvin Echols <calvin.echols@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"html/template"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/calvine/richerror/internal/templates"
	"github.com/calvine/richerror/internal/utilities"
	"github.com/spf13/cobra"
)

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

type generatorData struct {
	ErrorPkg string
	errorData
}

const (
	FlagErrorsDefinitionFile = "errorsDefinitionFile"
	FlagOutDir               = "outDir"
	FlagOutputErrorPkg       = "outputErrorPkg"
	FlagIncludeTags          = "includeTags"
	FlagExcludeTags          = "excludeTags"
	// FlagOutputCodePkg        = "outputCodePkg"
	// FlagTargetPackage = "targetPkg"
)

// generateCmd represents the generate command
var (
	errorsDefinitionFile string
	outDir               string
	outputErrorPkg       string
	includeTags          string
	excludeTags          string
	// outputCodePkg        string
	// targetPkg            string

	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generates error constructors and code constants.",
		Long:  ``,
		Run:   errorGenerator,
	}
)

func initGenerator() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command

	// This flags are persistent because at soom point other languages could be sub commands to this command.
	generateCmd.PersistentFlags().StringVarP(&errorsDefinitionFile, FlagErrorsDefinitionFile, "i", "", "The path to the errors definition file to use for error generation.")
	generateCmd.MarkPersistentFlagRequired(FlagErrorsDefinitionFile)
	generateCmd.PersistentFlags().StringVarP(&outDir, FlagOutDir, "o", ".", "The output path to place the generated files. Setting this to 'stdout' will print the generated files to stdout.")
	generateCmd.PersistentFlags().StringVarP(&outputErrorPkg, FlagOutputErrorPkg, "e", "errors", "The package to put at the top of the generated error files")
	generateCmd.PersistentFlags().StringVarP(&includeTags, FlagIncludeTags, "t", "", fmt.Sprintf("Specifies the errors to perform code generation on based on the tags associated with it in the error definion file. Multiple tags are seperated by commas. This is mutually exclusive with %s", FlagExcludeTags))
	generateCmd.PersistentFlags().StringVarP(&excludeTags, FlagExcludeTags, "x", "", fmt.Sprintf("Specifies the errors to exclude from code generation on based on the tags associated with it in the error definion file. Multiple tags are seperated by commas. This is mutually exclusive with %s", FlagIncludeTags))
	// generateCmd.Flags().StringVarP(&outputCodePkg, FlagOutputCodePkg, "c", "codes", "The package to put at the top of the generated error code files")
}

func errorGenerator(cmd *cobra.Command, args []string) {
	// fmt.Printf("%s - %s - %s", errorsDefinitionFile, outDir, outputErrorPkg)
	errorsDir := path.Join(outDir, strings.ToLower(outputErrorPkg))
	errorsDirExists, _ := utilities.DirExists(errorsDir)
	if !errorsDirExists {
		err := os.MkdirAll(errorsDir, os.ModePerm)
		if err != nil {
			panic(err.Error())
		}
	}
	// codesDir := path.Join(outDir, strings.ToLower(outputErrorPkg), strings.ToLower(outputCodePkg))
	funcMap := template.FuncMap{
		"ToUpper":            strings.ToUpper,
		"ToLower":            strings.ToLower,
		"UpperCaseFirstChar": utilities.UpperCaseFirstChar,
		"LowerCaseFirstChar": utilities.LowerCaseFirstChar,
	}
	errConstructorTemplate := template.Must(template.New("Error constructor template").Parse(templates.ErrorConstructorTemplate)).Funcs(funcMap)
	// errCodeTemplate := template.Must(template.New("Error code template").Parse(templates.ErrorCodeTemplate)).Funcs(funcMap)
	errDataSlice := make([]errorData, 0)
	jsonErrorDataFileData, err := ioutil.ReadFile(errorsDefinitionFile)
	if err != nil {
		errMsg := fmt.Sprintf("failed to open file %s - %s", errorsDefinitionFile, err.Error())
		panic(errMsg)
	}
	json.Unmarshal(jsonErrorDataFileData, &errDataSlice)
	if includeTags != "" {
		specificTags := strings.Split(includeTags, ",")
		fmt.Printf("Include tags specified. Filtering error definitions to only generate errors with the following tags: %s\n\n", includeTags)
		errDataSlice = getMatchingErrorsByTag(errDataSlice, specificTags, true)
	} else if excludeTags != "" {
		specificTags := strings.Split(excludeTags, ",")
		fmt.Printf("Exclude tags specified. Filtering error definitions to only generate errors without the following tags: %s\n\n", excludeTags)
		errDataSlice = getMatchingErrorsByTag(errDataSlice, specificTags, false)
	}
	fmt.Printf("generating %d errors.\n\n", len(errDataSlice))
	for _, data := range errDataSlice {
		genData := generatorData{outputErrorPkg, data}
		constructorBuffer := bytes.NewBufferString("")
		err := errConstructorTemplate.Execute(constructorBuffer, genData)
		if err != nil {
			fmt.Printf("failed to execute error constructor template: %s\n", err.Error())
			continue
		}
		errConstructorCode, err := format.Source(constructorBuffer.Bytes())
		if err != nil {
			fmt.Printf("%s", constructorBuffer)
			fmt.Printf("Failed to run format.Source on error code template: %s\n", err.Error())
			continue
		}

		// codeBuffer := bytes.NewBufferString("")
		// err = errCodeTemplate.Execute(codeBuffer, genData)
		// if err != nil {
		// 	fmt.Printf("failed to execute error code template: %s", err.Error())
		// 	continue
		// }
		// errCodeCode, err := format.Source([]byte(codeBuffer.String()))
		// if err != nil {
		// 	fmt.Printf("%s", codeBuffer)
		// 	fmt.Printf("Failed to run format.Source on error code template: %s", err.Error())
		// 	continue
		// }

		if outDir == "stdout" {
			fmt.Printf("\n\n************** %s Error Code **************\n\n", data.Code)
			fmt.Fprint(os.Stdout, string(errConstructorCode))
			fmt.Printf("\n\n****************************************************")
			// fmt.Printf("\n\n************** %s Error Code Code **************\n\n", data.Code)
			// fmt.Fprint(os.Stdout, string(errCodeCode))
			// fmt.Printf("\n\n*********************************************")
		} else {
			// emit files...
			fileName := fmt.Sprintf("%s.go", strings.ToLower(data.Code))
			errConstructorFilePath := path.Join(errorsDir, fileName)
			fmt.Printf("Generating code for error code: %s -> %s\n", data.Code, errConstructorFilePath)
			err = ioutil.WriteFile(errConstructorFilePath, errConstructorCode, fs.ModePerm)
			if err != nil {
				fmt.Printf("Failed to write file %s for err constructor for code %s - %s\n\n\n", errConstructorFilePath, data.Code, err.Error())
				continue
			}
			// errCodeFilePath := path.Join(codesDir, fileName)
			// err = ioutil.WriteFile(errCodeFilePath, errCodeCode, fs.ModePerm)
			// if err != nil {
			// 	fmt.Printf("Failed to write file %s for err code for code %s", errCodeFilePath, data.Code)
			// 	continue
			// }
		}
	}
}

func getMatchingErrorsByTag(data []errorData, tags []string, isInclude bool) []errorData {
	matchingErrors := make([]errorData, 0)
	for _, errDefinition := range data {
		hasMatchingTag := false
		var firstMatchedTag string
		for _, errTag := range errDefinition.Tags {
			errTag = strings.TrimSpace(strings.ToLower(errTag))
			for _, cliTag := range tags {
				cliTag = strings.TrimSpace(strings.ToLower(cliTag))
				if errTag == cliTag {
					firstMatchedTag = errTag
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				break
			}
		}
		if isInclude && hasMatchingTag {
			fmt.Printf("Added for generation: Error '%s' has matching tag '%s'\n", errDefinition.Code, firstMatchedTag)
			matchingErrors = append(matchingErrors, errDefinition)
		} else if !isInclude && !hasMatchingTag {
			fmt.Printf("Added for generation: Error '%s' does not have tag '%s'\n", errDefinition.Code, firstMatchedTag)
			matchingErrors = append(matchingErrors, errDefinition)
		}
	}
	fmt.Printf("\n%d errors matched the tags provided.\n\n", len(matchingErrors))
	return matchingErrors
}
