package audio

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type Downloader struct {
	IgnoreChars []string
	AudioDir    string
}

// download audio file from google text-to-speech api if it doesn't exist in cache dir.
// we also store a sentenceAndDialogOnlyDir to create audio loops for which we want to exclude words and chars.
func (p *Downloader) Fetch(ctx context.Context, query, filename string, isSentenceOrDialog bool) (string, error) {
	filename = filename + ".mp3"
	if contains(p.IgnoreChars, query) {
		return filename, nil
	}
	if err := os.MkdirAll(p.AudioDir, os.ModePerm); err != nil {
		return "", err
	}
	sentenceAndDialogOnlyDir := filepath.Join(p.AudioDir, "sentences_and_dialogs")
	if err := os.MkdirAll(sentenceAndDialogOnlyDir, os.ModePerm); err != nil {
		return "", err
	}
	lessonPath := filepath.Join(p.AudioDir, filename)
	cachePath := filepath.Join(p.AudioDir, "..", "..", "..", "audio", filename)
	sentenceAndDialogOnlyPath := filepath.Join(sentenceAndDialogOnlyDir, filename)

	// copy file from cache to lesson dir and to sentenceAndDialogOnlyDir
	if _, err := os.Stat(cachePath); err == nil {
		var hasErr bool
		if err := copyFileContents(cachePath, lessonPath); err != nil {
			hasErr = true
			fmt.Printf("error copying cache file for query %s: %v\n", query, err)
		}
		if isSentenceOrDialog {
			if err := copyFileContents(cachePath, sentenceAndDialogOnlyPath); err != nil {
				hasErr = true
				fmt.Printf("error copying cache file for query %s: %v\n", query, err)
			}
		}
		if !hasErr {
			return filename, nil
		}
	}

	time.Sleep(100 * time.Millisecond)
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: query},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "cmn-CN",
			Name:         "cmn-CN-Wavenet-C",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return "", err
	}

	// The resp's AudioContent is binary.
	err = ioutil.WriteFile(lessonPath, resp.AudioContent, os.ModePerm)
	if err != nil {
		return "", err
	}
	// for creating audio loops from sentences and dialogs only.
	if isSentenceOrDialog {
		err = ioutil.WriteFile(sentenceAndDialogOnlyPath, resp.AudioContent, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	err = ioutil.WriteFile(cachePath, resp.AudioContent, os.ModePerm)
	if err != nil {
		return "", err
	}

	fmt.Printf("%v\n", query)
	if isSentenceOrDialog {
		fmt.Printf("audio content written to files:\n%s\n%s\n%s\n", lessonPath, cachePath, sentenceAndDialogOnlyPath)
	} else {
		fmt.Printf("audio content written to files:\n%s\n%s\n", lessonPath, cachePath)
	}
	return filename, nil
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
