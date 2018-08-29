package ocr

import (
	"testing"
)

func TestNewIdRecognition(t *testing.T) {
	file := "/home/lezh1k/test_data/idmatch_data/n8.jpg"
	template := "KG idcard new"
	prev := "/tmp"
	result, path := Recognize(file, "", template, prev)
	t.Log(result, path)
	t.Fail()
}
