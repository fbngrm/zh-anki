package decomposition

import (
	"fmt"
	"os"
	"strings"

	"github.com/fbngrm/zh/lib/cedict"
	"github.com/fbngrm/zh/lib/hanzi"
	"github.com/fbngrm/zh/pkg/cjkvi"
	"github.com/fbngrm/zh/pkg/finder"
	"github.com/fbngrm/zh/pkg/frequency"
	"github.com/fbngrm/zh/pkg/kangxi"
	"github.com/fbngrm/zh/pkg/search"
)

const idsSrc = "./pkg/cjkvi/ids.txt"
const cedictSrc = "./pkg/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const wordFrequencySrc = "./pkg/frequency/global_wordfreq.release_UTF-8.txt"

type Decomposer struct {
	hanziDecomposer *hanzi.Decomposer
}

func NewDecomposer() *Decomposer {
	dict, err := cedict.NewDict(cedictSrc)

	idsDecomposer, err := cjkvi.NewIDSDecomposer(idsSrc)
	if err != nil {
		fmt.Printf("could not initialize ids decompose: %v\n", err)
		os.Exit(1)
	}

	return &Decomposer{
		hanziDecomposer: hanzi.NewDecomposer(
			dict,
			kangxi.NewDict(),
			search.NewSearcher(finder.NewFinder(dict)),
			idsDecomposer,
			nil,
			frequency.NewWordIndex(wordFrequencySrc),
		),
	}
}

func (d *Decomposer) Decompose(word string) hanzi.Hanzi {
	decomposition, err := d.hanziDecomposer.Decompose(word, 3, 0)
	if err != nil {
		fmt.Printf("error decomposing %s: %v\n", word, err)
		f, err := os.OpenFile("./decomp_err_log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Printf("error opening error log %s: %v\n", word, err)
		}
		defer f.Close()
		if _, err = f.WriteString(word + "\n"); err != nil {
			fmt.Printf("error writing error log %s: %v\n", word, err)
		}
		return hanzi.Hanzi{}
	}
	if len(decomposition.Errs) != 0 {
		for _, e := range decomposition.Errs {
			fmt.Printf("errors decomposing %s: %v\n", word, e)
		}
		return hanzi.Hanzi{}
	}
	if len(decomposition.Hanzi) != 1 {
		fmt.Printf("expect exactly 1 decomposition for word: %s", word)
		return hanzi.Hanzi{}
	}
	return *decomposition.Hanzi[0]
}

func GetKangxi(h hanzi.Hanzi) []string {
	components := []string{}
	for _, c := range h.ComponentsDecompositions {
		if c.IsKangxi {
			// we limit the definitions to 3 per cedict entry
			componentDefintions := make([]string, 0)
			for _, definition := range c.Definitions {
				definitions := strings.Split(definition, ",")
				if len(definitions) > 3 {
					definitions = definitions[:3]
				}
				componentDefintions = append(componentDefintions, definitions...)
			}
			component := c.Ideograph
			component += " = "
			component += strings.Join(componentDefintions, ", ")
			components = append(components, component)
		}
	}
	return components
}

func GetComponents(h hanzi.Hanzi) []string {
	components := []string{}
	for _, c := range h.ComponentsDecompositions {
		componentDefintions := make([]string, 0)
		for _, definition := range c.Definitions {
			definitions := strings.Split(definition, ",")
			componentDefintions = append(componentDefintions, definitions...)
		}
		component := c.Ideograph
		component += " = "
		component += strings.Join(componentDefintions, ", ")
		components = append(components, component)
	}
	return components
}
