package ocr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

//feel free to add something here
var imageFormats = [...]string{
	".jpg",
	".jpeg",
	".png",
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func isImage(ext string) bool {
	for _, f := range imageFormats {
		if strings.Compare(ext, f) == 0 {
			return true
		}
	}
	return false
}

func compareJSONAndOcr(json, ocr map[string]interface{}) float64 {
	sumErr := 0
	sumStr := 0
	for k, v := range json {
		vv := v.(string)
		vv = strings.ToUpper(vv)
		sumStr += len(vv)
		ov, ok := ocr[k]
		if !ok {
			sumErr += len(vv)
			ov = ""
		}
		ovv := ov.(string)
		ovv = strings.ToUpper(ovv)

		dd := levenshtein.DistanceForStrings([]rune(vv), []rune(ovv), levenshtein.DefaultOptions)
		fmt.Printf("%s -> %s == %d\n", vv, ovv, dd)
		sumErr += dd
	}
	fmt.Printf("SUMMARY : %d\n", sumErr)
	return float64(sumErr) / float64(sumStr)
}

//CheckSolution takes folder with pictures and json-files
//and tries to recognize text on pictures. Then with levenstein distance
//it tries to calculate match percentage
func CheckSolution(folderPath, flagTemplate string) error {

	fmt.Println(folderPath)
	filepath.Walk(folderPath,
		func(path string, info os.FileInfo, err error) error {
			ext := filepath.Ext(path)
			if info.IsDir() || !isImage(ext) {
				return nil
			}

			fileNameWOExt := strings.TrimSuffix(info.Name(), ext)
			jsonFileName := folderPath + fileNameWOExt + ".json"

			if !fileExists(jsonFileName) {
				return nil
			}

			jsonFileData, err := ioutil.ReadFile(jsonFileName)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}

			//todo move these functions to some map with template as key
			var jsonInfo map[string]interface{}
			err = json.Unmarshal(jsonFileData, &jsonInfo)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}

			ocrInfo, _ := Recognize(path, flagTemplate, "")
			fmt.Println((1.0 - compareJSONAndOcr(jsonInfo, ocrInfo)) * 100.0)
			return nil
		})
	return nil
}
