package translate

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Translations struct {
	dict          map[string]string
	charsToRemove []string
}

func New(path string, charsToRemove []string) (*Translations, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open translations file: %w", err)
	}
	var dict map[string]string
	if err := yaml.Unmarshal(b, &dict); err != nil {
		return nil, fmt.Errorf("could not unmarshal translations file: %w", err)
	}
	return &Translations{
		dict:          dict,
		charsToRemove: charsToRemove,
	}, nil
}

func (t *Translations) Lookup(s string) string {
	return t.dict[removeChars(s, t.charsToRemove)]
}

// removeChars removes all characters from the given string that are present in the charsToRemove.
func removeChars(input string, charsToRemove []string) string {
	result := strings.Builder{}
	for _, char := range input {
		if !contains(charsToRemove, string(char)) {
			result.WriteString(string(char))
		}
	}
	return result.String()
}

// contains checks if a slice of strings contains a specific string.
func contains(slice []string, char string) bool {
	for _, c := range slice {
		if c == char {
			return true
		}
	}
	return false
}
