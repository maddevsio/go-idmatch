package utils

import (
	"regexp"
	"strings"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"
)

func gender(gender string) string {
	if strings.ContainsAny(gender, "m M э Э м М 3") {
		return "Э"
	} else if strings.ContainsAny(gender, "f F а А ж Ж") {
		return "А"
	}
	return ""
}

func Sanitize(documentMap map[string]interface{}, card templates.Card) {
	regex := "[^а-яa-zА-ЯA-Z0-9]+"

	for _, v := range card.Structure {
		if documentMap[v.Field] == nil {
			continue
		}
		text := documentMap[v.Field].(string)

		switch v.Type {
		case "cyrillic":
			regex = "[^а-яА-Я]+"
		case "latin":
			regex = "[^a-zA-Z]+"
		case "number":
			regex = "[^0-9]+"
		case "gender":
			text = gender(text)
			regex = "[^а-яА-Я]+"
		}

		reg, err := regexp.Compile(regex)
		if err != nil {
			log.Print(log.ErrorLevel, err.Error())
		}
		clearText := reg.ReplaceAllString(text, "")
		if len(clearText) == 0 {
			clearText = "(recognition failed)"
		} else if text != clearText {
			clearText += " (?)"
		}
		documentMap[v.Field] = clearText
	}
}
