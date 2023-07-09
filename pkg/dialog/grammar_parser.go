package dialog

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func loadGrammarYAML(path string) (Grammar, error) {
	d, err := os.ReadFile(path)
	if err != nil {
		return Grammar{}, fmt.Errorf("error reading grammar file: %v", err)
	}

	var g Grammar
	if err = yaml.Unmarshal([]byte(d), &g); err != nil {
		return Grammar{}, fmt.Errorf("error unmarshalling grammar file: %v", err)
	}

	return g, nil
}

// func loadGrammarLines(path string) ([]Grammar, error) {
// 	content, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading grammar file: %v", err)
// 	}

// 	parts := strings.Split(string(content), "===\n")

// 	var grammar []Grammar
// 	for _, part := range parts {
// 		splits := strings.Split(part, "---\n")

// 		if len(splits) == 0 {
// 			continue
// 		}

// 		fmt.Println(len(splits))
// 		fmt.Println(splits)

// 		if len(splits) != 5 {
// 			return nil, errors.New("invalid file format")
// 		}

// 		fileSplit := Grammar{
// 			Head:               strings.TrimSpace(splits[0]),
// 			Description:        strings.TrimSpace(splits[1]),
// 			Structure:          strings.TrimSpace(splits[2]),
// 			Examples:           strings.Split(strings.TrimSpace(splits[3]), "\n"),
// 			ExampleDescription: strings.TrimSpace(splits[4]),
// 		}
// 		grammar = append(grammar, fileSplit)
// 	}
// 	return grammar, nil
// }
