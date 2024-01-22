package anki

import (
	"fmt"
	"os"

	"github.com/fbngrm/zh-anki/pkg/template"
	"gopkg.in/yaml.v2"
)

type Exporter struct {
	TmplProcessor *template.Processor
}

func (e *Exporter) CreateOrAppendAnkiCards(a any, templateName, outPath string) {
	text, err := e.TmplProcessor.Fill(a, templateName)
	if err != nil {
		fmt.Printf("could not fill template file: %v\n", err)
		os.Exit(1)
	}
	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("could not open anki cards file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		fmt.Printf("could not append to anki cards file: %v\n", err)
	}
}

func (e *Exporter) WriteYAMLFile(i any, path string) {
	data, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("could not marshal interface: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("could not write file: %v", err)
		os.Exit(1)
	}
}
