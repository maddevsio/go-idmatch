package utils

import (
	"regexp"
	"strings"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"
)

const ErrorMessage = "(recognition failed)"

func gender(gender string) string {
	if strings.ContainsAny(gender, "m M э Э м М 3 9 5") {
		return "Э"
	} else if strings.ContainsAny(gender, "f F а А ж Ж") {
		return "А"
	}
	return ""
}

func Sanitize(data []templates.Field) {
	regex := "[^а-яa-zА-ЯA-Z0-9№ ]+"

	for i, field := range data {
		switch field.Type {
		case "cyrillic":
			regex = "[^а-яА-Я№ ]+"
		case "latin":
			regex = "[^a-zA-Z ]+"
		case "number":
			regex = "[^0-9]+"
		case "gender":
			data[i].Text = gender(field.Text)
			regex = "[^а-яА-Я]$"
		}

		if n := strings.Index(data[i].Text, "\n"); n > 0 {
			data[i].Text = data[i].Text[:n]
		}
		reg, err := regexp.Compile(regex)
		if err != nil {
			log.Print(log.ErrorLevel, err.Error())
		}
		clearText := reg.ReplaceAllString(data[i].Text, "")
		if len(clearText) == 0 {
			clearText = ErrorMessage
			continue
		}
		if field.Length != 0 && len(clearText) > field.Length {
			clearText = clearText[len(clearText)-field.Length:]
		}
		if field.Prefix != "" {
			clearText = field.Prefix + clearText
		}
		// else if text != clearText {
		// 	clearText += " (?)"
		// }
		data[i].Text = strings.ToUpper(clearText)
	}
}
