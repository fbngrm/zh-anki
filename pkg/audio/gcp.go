package audio

import (
	"context"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"golang.org/x/exp/slog"
)

type GCPClient struct {
	IgnoreChars []string
	AudioDir    string
}

// we support 4 different voices only
var voices = []*texttospeechpb.VoiceSelectionParams{
	{
		LanguageCode: "cmn-CN",
		Name:         "cmn-CN-Wavenet-C",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
	},
	{
		LanguageCode: "cmn-CN",
		Name:         "cmn-CN-Wavenet-A",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
	},
	{
		LanguageCode: "cmn-CN",
		Name:         "cmn-TW-Wavenet-C",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
	},
	{
		LanguageCode: "cmn-CN",
		Name:         "cmn-TW-Wavenet-A",
		SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
	},
}

func (g *GCPClient) GetVoices(
	speakers map[string]struct{},
) map[string]*texttospeechpb.VoiceSelectionParams {
	v := make(map[string]*texttospeechpb.VoiceSelectionParams)
	var i int
	for speaker := range speakers {
		v[speaker] = voices[i]
		i++
	}
	return v
}

// download audio file from google text-to-speech api if it doesn't exist in cache dir.
// we also store a sentenceAndDialogOnlyDir to create audio loops for which we want to exclude words and chars.
func (g *GCPClient) Fetch(ctx context.Context, query, filename string, isSentenceOrDialog bool) error {
	if contains(g.IgnoreChars, query) {
		return nil
	}
	if err := os.MkdirAll(g.AudioDir, os.ModePerm); err != nil {
		return err
	}
	lessonPath := filepath.Join(g.AudioDir, filename)

	resp, err := fetch(ctx, query, nil)
	if err != nil {
		return err
	}

	// the resp's AudioContent is binary.
	err = ioutil.WriteFile(lessonPath, resp.AudioContent, os.ModePerm)
	if err != nil {
		return err
	}
	slog.Debug("download GCP audio", "path", lessonPath)
	return nil
}

// uses a random voice if param voice is nil
func fetch(ctx context.Context, query string, voice *texttospeechpb.VoiceSelectionParams) (*texttospeechpb.SynthesizeSpeechResponse, error) {
	time.Sleep(100 * time.Millisecond)
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	if voice == nil {
		rand.Seed(time.Now().UnixNano()) // initialize global pseudo random generator
		voice = voices[rand.Intn(len(voices))]
	}
	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: query},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: voice,
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:  1,
		},
	}
	return client.SynthesizeSpeech(ctx, &req)
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
