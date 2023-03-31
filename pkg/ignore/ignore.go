package ignore

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Ignored map[string]struct{}

func (i Ignored) Update(s string) {
	i[s] = struct{}{}
}

func Load(path string) Ignored {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("could not open ignore file: %v", err)
		os.Exit(1)
	}
	var i Ignored
	if err := yaml.Unmarshal(b, &i); err != nil {
		fmt.Printf("could not unmarshal ignore file: %v", err)
		os.Exit(1)
	}
	return i
}

func (i Ignored) Write(path string) {
	data, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("could not marshal ignore file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write ignore file: %v", err)
		os.Exit(1)
	}
}

func writeToFile(i any, path string) {
	data, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("could not marshal ignore file: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write ignore file: %v", err)
		os.Exit(1)
	}
}
