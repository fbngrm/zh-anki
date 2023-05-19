package decomposition

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/fbngrm/zh/lib/cedict"
	"github.com/fbngrm/zh/lib/hanzi"
	"github.com/fbngrm/zh/pkg/cjkvi"
	"github.com/fbngrm/zh/pkg/finder"
	"github.com/fbngrm/zh/pkg/frequency"
	"github.com/fbngrm/zh/pkg/kangxi"
	"github.com/fbngrm/zh/pkg/search"
)

const idsSrc = "/home/f/work/src/github.com/fbngrm/zh-anki/lib/cjkvi/ids.txt"
const cedictSrc = "/home/f/work/src/github.com/fbngrm/zh/lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const wordFrequencySrc = "../../lib/word_frequencies/global_wordfreq.release_UTF-8.txt"

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

	// we provide a word frequency index which needs to be initialized before first use.
	frequencyIndex := frequency.NewWordIndex(wordFrequencySrc)

	return &Decomposer{
		hanziDecomposer: hanzi.NewDecomposer(
			dict,
			kangxi.NewDict(),
			search.NewSearcher(finder.NewFinder(dict)),
			idsDecomposer,
			nil,
			frequencyIndex,
		),
	}
}

func (d *Decomposer) Decompose(word string) {
	decomposition, err := d.hanziDecomposer.Decompose(word, 3, 0)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error decomposing %s: %v\n", word, err))
	}
	if len(decomposition.Errs) != 0 {
		for _, e := range decomposition.Errs {
			os.Stderr.WriteString(fmt.Sprintf("errors decomposing %s: %v\n", word, e))
		}
	}
	if len(decomposition.Hanzi) != 1 {
		os.Stderr.WriteString(fmt.Sprintf("expect exactly 1 decomposition for word: %s", word))
		os.Exit(1)
	}
	spew.Dump(decomposition)
	// spew.Dump(decomposition)
}
