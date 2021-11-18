package utilities

import "github.com/calvine/richerror/internal/cmd/models"

func GetDataItemImportMap(items []models.DataItem) []string {
	uniqueImportsMap := make(map[string]bool)
	uniqueImports := make([]string, 0)
	for _, i := range items {
		if i.ImportPath != "" {
			uniqueImportsMap[i.ImportPath] = true
		}
	}
	for k := range uniqueImportsMap {
		uniqueImports = append(uniqueImports, k)
	}
	return uniqueImports
}
