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
	FlagTags                 = "tags"
	// FlagOutputCodePkg        = "outputCodePkg"
	// FlagTargetPackage = "targetPkg"
)

// generateCmd represents the generate command
var (
	errorsDefinitionFile string
	outDir               string
	outputErrorPkg       string
	tags                 string
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
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	generateCmd.Flags().StringVarP(&errorsDefinitionFile, FlagErrorsDefinitionFile, "i", "", "The path to the errors definition file to use for error generation.")
	generateCmd.MarkFlagRequired(FlagErrorsDefinitionFile)
	generateCmd.Flags().StringVarP(&outDir, FlagOutDir, "o", ".", "The output path to place the generated files. Setting this to 'stdout' will print the generated files to stdout.")
	generateCmd.Flags().StringVarP(&outputErrorPkg, FlagOutputErrorPkg, "e", "errors", "The package to put at the top of the generated error files")
	generateCmd.Flags().StringVarP(&tags, FlagTags, "t", "", "Specifies the errors to perform code generation on based on the tags associated with it in the error definion file. Multiple tags are seperated by commas")
	// generateCmd.Flags().StringVarP(&outputCodePkg, FlagOutputCodePkg, "c", "codes", "The package to put at the top of the generated error code files")
}

func errorGenerator(cmd *cobra.Command, args []string) {
	// fmt.Printf("%s - %s - %s", errorsDefinitionFile, outDir, outputErrorPkg)
	errorsDir := path.Join(outDir, strings.ToLower(outputErrorPkg))
	errorsDirExists, _ := utilities.DirExists(errorsDir)
	specificTags := strings.Split(tags, ",")
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
	if len(specificTags) > 0 {
		fmt.Printf("Tags specified. Filtering error definitions to only generate errors with the following tags: %s\n\n", tags)
		errDataSlice = getMatchingErrorsByTag(errDataSlice, specificTags)
	}
	fmt.Printf("generating %d errors.", len(errDataSlice))
	for _, data := range errDataSlice {
		genData := generatorData{outputErrorPkg, data}
		constructorBuffer := bytes.NewBufferString("")
		err := errConstructorTemplate.Execute(constructorBuffer, genData)
		if err != nil {
			fmt.Printf("failed to execute error constructor template: %s", err.Error())
			continue
		}
		errConstructorCode, err := format.Source(constructorBuffer.Bytes())
		if err != nil {
			fmt.Printf("%s", constructorBuffer)
			fmt.Printf("Failed to run format.Source on error code template: %s", err.Error())
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

func getMatchingErrorsByTag(data []errorData, tags []string) []errorData {
	matchingTagErrors := make([]errorData, 0)
	for _, errDefinition := range data {
		hasMatchingTag := false
		for _, errTag := range errDefinition.Tags {
			for _, cliTag := range tags {
				if errTag == cliTag {
					fmt.Printf("Error '%s' has matching tag '%s'\n", errDefinition.Code, errTag)
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				matchingTagErrors = append(matchingTagErrors, errDefinition)
				break
			}
		}
	}
	fmt.Printf("\n%d errors matched the tags provided.\n\n", len(matchingTagErrors))
	return matchingTagErrors
}
