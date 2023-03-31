package translate

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Translations map[string]string

func (t Translations) Update(ch, en string) {
	t[ch] = en
}

func Load(path string) Translations {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("could not open translations file: %v", err)
		os.Exit(1)
	}
	var t Translations
	if err := yaml.Unmarshal(b, &t); err != nil {
		fmt.Printf("could not unmarshal translations file: %v", err)
		os.Exit(1)
	}
	return t
}

func (t Translations) Write(path string) {
	data, err := yaml.Marshal(t)
	if err != nil {
		fmt.Printf("could not marshal translations file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write translations file: %v", err)
		os.Exit(1)
	}
}
