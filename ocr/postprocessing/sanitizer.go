package postprocessing

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	translit "github.com/gen1us2k/go-translit"
	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

const (
	dateLayout = "02012006"
)

func Sanitize(data []templates.Side) map[string]interface{} {
	result := make(map[string]interface{})
	for i, side := range data {
		for j, field := range side.Structure {
			if _, ok := result[field.Name]; ok {
				continue
			}

			if field.Minlength == 0 {
				field.Minlength = 1
			}

			log.Print(log.DebugLevel, "***"+field.Name+"***")
			log.Print(log.DebugLevel, fmt.Sprintf("%s %s", field.Text, field.Type))
			regex := fieldRegex(field)
			if n := strings.Index(field.Text, "\n"); n > 0 && !field.Multiline {
				field.Text = field.Text[:n]
			}
			reg, err := regexp.Compile(regex)
			if err != nil {
				log.Print(log.ErrorLevel, err.Error())
				continue
			}
			clearText := strings.TrimSpace(reg.ReplaceAllString(field.Text, ""))
			if (field.Length != 0 && len(clearText) != field.Length) || len(clearText) < field.Minlength {
				log.Print(log.DebugLevel, fmt.Sprintf("Length mismatch: %s(%d), length: %d, min length: %d\n", clearText, len(clearText), field.Length, field.Minlength))
				log.Print(log.DebugLevel, "***END***")
				continue
			}

			fmt.Println(clearText, field.Minlength)

			if len(field.Fragment) != 0 {
				// fmt.Println("FRAGMENT: ", field.Type, clearText, n)
				clearText = getFragment(clearText, field.Fragment)
				fmt.Println("CCC: ", clearText)
				if len(clearText) == 0 {
					log.Print(log.DebugLevel, "Fragment length mismatch")
					log.Print(log.DebugLevel, "***END***")
					continue
				}
			}
			clearText = field.Prefix + clearText

			if len(field.Options) != 0 {
				min := 10
				for _, option := range field.Options {
					if m := distance(clearText, option); m < min {
						min = m
						field.Text = option
					}
				}
				clearText = field.Text
			}

			switch field.Type {
			case "date":
				if !isDate(clearText) {
					continue
				}
				clearText = clearText[:2] + "." + clearText[2:4] + "." + clearText[4:8]
			case "gender":
				clearText = gender(clearText)
			}

			if len(field.Subfield.Fields) != 0 {
				if fields := strings.Split(clearText, field.Subfield.Delimeter); len(fields) >= len(field.Subfield.Fields) {
					for k, v := range field.Subfield.Fields {
						if _, ok := result[v]; !ok {
							result[v] = fields[k]
							if field.Transliterate {
								result[v] = translit.Translit(fields[k])
							}
						}
					}
				}
			}

			if len(clearText) != 0 && len(field.Name) != 0 {
				log.Print(log.DebugLevel, field.Name+" : "+clearText)
				log.Print(log.DebugLevel, "***SUCCESS***")
				data[i].Structure[j].Text = clearText
				result[field.Name] = clearText
			}
		}
	}
	return result
}

func fieldRegex(field templates.Field) string {
	regex := "[^А-Я ^A-Z ^0-9 ^. ^-]+"
	switch field.Type {
	case "string":
		regex = "[^А-Я ^A-Z ]+"
	case "date", "number":
		regex = "[^0-9]+"
	case "custom":
		regex = field.Regex
	}
	return regex
}

func gender(gender string) string {
	if strings.ContainsAny(gender, "m M э Э м М 3 9 5 2") {
		return "Э"
	} else if strings.ContainsAny(gender, "f F а А ж Ж 1") {
		return "А"
	}
	return ""
}

func distance(source, target string) int {
	return levenshtein.DistanceForStrings([]rune(source), []rune(target), levenshtein.DefaultOptions)
}

func isDate(value string) bool {
	t, err := time.Parse(dateLayout, value)
	if err != nil {
		// log.Print(log.WarnLevel, err.Error())
		return false
	}
	if t.Format(dateLayout) != value {
		// log.Print(log.WarnLevel, "Date format error: "+value)
		return false
	}
	return true
}

func getFragment(text, fragment string) string {
	pos := strings.Split(fragment, "-")
	begin, err := strconv.Atoi(pos[0])
	if err != nil {
		return ""
	}
	end, err := strconv.Atoi(pos[1])
	if err != nil {
		return ""
	}
	if len(pos) != 2 || len(text) < end {
		return ""
	}
	return text[begin:end]
}
