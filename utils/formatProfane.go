package utils

import (
	"strings"
	"errors"
)
func FormatProfane(originalText string, sanitizedText string) (string, error) {
	words := strings.Split(sanitizedText, " ")
	wordsToReturn := strings.Split(originalText, " ")
	if len(words) != len(wordsToReturn) {
		return "", errors.New("words whit diferent sizes, cant format it")
	}
	for i := 0; i < len(words); i++ {
		if strings.Contains(words[i], "****") {
			wordsToReturn[i] = words[i]
		}
	}
	return strings.Join(wordsToReturn, " "), nil
}