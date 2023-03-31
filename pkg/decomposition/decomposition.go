package decomposition

// func addDecompositionsToWords(pinyinIndex Pinyin, words []Word) []Word {
// 	dict, err := cedict.NewDict(cedictSrc)

// 	idsDecomposer, err := cjkvi.NewIDSDecomposer(idsSrc)
// 	if err != nil {
// 		fmt.Printf("could not initialize ids decompose: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// we provide a word frequency index which needs to be initialized before first use.
// 	frequencyIndex := frequency.NewWordIndex(wordFrequencySrc)

// 	decomposer := hanzi.NewDecomposer(
// 		dict,
// 		kangxi.NewDict(),
// 		search.NewSearcher(finder.NewFinder(dict)),
// 		idsDecomposer,
// 		nil,
// 		frequencyIndex,
// 	)

// 	for _, word := range words {
// 		decomposition, err := decomposer.Decompose(word.Chinese, 3, 0)
// 		if err != nil {
// 			os.Stderr.WriteString(fmt.Sprintf("error decomposing %s: %v\n", word.Chinese, err))
// 		}
// 		if len(decomposition.Errs) != 0 {
// 			for _, e := range decomposition.Errs {
// 				os.Stderr.WriteString(fmt.Sprintf("errors decomposing %s: %v\n", word.Chinese, e))
// 			}
// 		}
// 		if len(decomposition.Hanzi) != 1 {
// 			os.Stderr.WriteString(fmt.Sprintf("expect exactly 1 decomposition for word: %s", word.Chinese))
// 			os.Exit(1)
// 		}
// 		spew.Dump(decomposition.Hanzi[0].Readings)
// 		// spew.Dump(decomposition)
// 	}
// 	return words
// }
