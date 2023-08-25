package cmds

import (
	"fmt"
	"os"
	"path"
)

func EepromToFiles(outputDir string, moduleName string, data [][]byte) {

	folderPath := path.Join(outputDir, moduleName)

	os.MkdirAll(folderPath, os.ModePerm)

	for timestamp, value := range data {
		filePath := path.Join(folderPath, fmt.Sprintf("%s-%09d", moduleName, timestamp))
		os.Create(filePath)

		if err := os.WriteFile(filePath, value, os.ModePerm); err != nil {
			fmt.Printf("Error while saving file %s\n", filePath)
			panic(1)
		}
	}
}
