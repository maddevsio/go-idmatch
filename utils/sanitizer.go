package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"
	"github.com/texttheater/golang-levenshtein/levenshtein"
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

func Sanitize(data []templates.Side) {
	regex := "[^а-яa-zА-ЯA-Z0-9№ ]+"

	for i, side := range data {
		for j, field := range side.Structure {
			switch field.Type {
			case "string":
				regex = "[^а-яА-Я ^a-zA-Z ]+"
			case "number":
				regex = "[^0-9]+"
			case "gender":
				data[i].Structure[j].Text = gender(field.Text)
				regex = "[^а-яА-Я]$"
			case "date":
				regex = "[^0-9]+"
			}

			if n := strings.Index(data[i].Structure[j].Text, "\n"); n > 0 {
				data[i].Structure[j].Text = data[i].Structure[j].Text[:n]
			}
			reg, err := regexp.Compile(regex)
			if err != nil {
				log.Print(log.ErrorLevel, err.Error())
			}
			clearText := reg.ReplaceAllString(data[i].Structure[j].Text, "")
			if len(clearText) == 0 {
				clearText = ErrorMessage
				continue
			}
			if field.Length != 0 {
				if len(clearText) > field.Length {
					clearText = clearText[len(clearText)-field.Length:]
				} else if len(clearText) < field.Length {
					clearText = padLeft(clearText, field.Length)
				}
			}
			if field.Prefix != "" {
				clearText = field.Prefix + clearText
			}
			if field.Type == "date" {
				clearText = clearText[:2] + "." + clearText[2:4] + "." + clearText[4:8]
			}

			min := 10
			for _, option := range field.Options {
				if m := distance(clearText, option); m < min {
					min = m
					clearText = option
				}
			}
			// if len(field.Validate) != 0 {
			// 	fmt.Println(field.Validate)
			// }
			// else if text != clearText {
			// 	clearText += " (?)"
			// }
			fmt.Println(clearText)
			data[i].Structure[j].Text = strings.ToUpper(clearText)
		}
	}
}

func padLeft(str string, lenght int) string {
	for {
		str = "0" + str
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

func distance(source, target string) int {
	return levenshtein.DistanceForStrings([]rune(source), []rune(target), levenshtein.DefaultOptions)
}
