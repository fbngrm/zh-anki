package translate

import (
	"context"
	"fmt"
	"strings"

	google_translate "cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

// func translateWords(words []Word, t translate.Translations) []Word {
// 	var translated []Word
// 	for _, word := range words {
// 		translation, ok := t[word.Chinese]
// 		if !ok {
// 			var err error
// 			translation, err = translate.Translate("en-US", word.Chinese)
// 			if err != nil {
// 				log.Fatalf("could not translate word \"%s\": %v", word.Chinese, err)
// 			}
// 		}
// 		word.English = translation
// 		t.Update(word.Chinese, word.English)

//			translated = append(translated, word)
//		}
//		return translated
//	}

func Translate(targetLanguage, text string) (string, error) {
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %v", err)
	}

	client, err := google_translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	fmt.Printf("translate: %s...\n", text)
	resp, err := client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("translate: %v", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("translate returned empty response to text: %s", text)
	}
	trans := resp[0].Text
	trans = strings.ReplaceAll(trans, "&#39;", "'")
	fmt.Println(trans)
	return trans, nil
}
