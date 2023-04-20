package audio

import (
	"context"
	"fmt"
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

func (p *Downloader) Fetch(ctx context.Context, query, filename string) (string, error) {
	filename = filename + ".mp3"
	if contains(p.IgnoreChars, query) {
		return filename, nil
	}
	if err := os.MkdirAll(p.AudioDir, os.ModePerm); err != nil {
		return "", err
	}
	path := filepath.Join(p.AudioDir, filename)
	globalPath := filepath.Join(p.AudioDir, "..", "..", "..", "audio", filename)

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("audio file exists: %s\n", path)
		return filename, nil
	}
	if _, err := os.Stat(globalPath); err == nil {
		fmt.Printf("audio file exists: %s\n", globalPath)
		return filename, nil
	}

	time.Sleep(1 * time.Second)
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
	err = ioutil.WriteFile(path, resp.AudioContent, os.ModePerm)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(globalPath, resp.AudioContent, os.ModePerm)
	if err != nil {
		return "", err
	}

	fmt.Printf("%v\n", query)
	fmt.Printf("audio content written to file: %v\n", path)
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
