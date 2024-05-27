package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// speed
const rate = "0.7"

type AzureClient struct {
	endpoint    string
	apiKey      string
	AudioDir    string
	ignoreChars []string
}

func NewAzureClient(apiKey, audioDir string, ignoreChars []string) *AzureClient {
	return &AzureClient{
		endpoint:    "https://germanywestcentral.tts.speech.microsoft.com/cognitiveservices/v1",
		apiKey:      apiKey,
		AudioDir:    audioDir,
		ignoreChars: ignoreChars,
	}
}

// we support 4 different voices only
var Voices = []string{
	"zh-CN-YunzeNeural",
	"zh-CN-XiaobeiNeural",
	"zh-CN-YunjieNeural",
	"zh-CN-XiaochenNeural",
	"zh-CN-JunjianNeural",
	"zh-CN-XiaohanNeural",
}

func (c *AzureClient) GetRandomVoice() string {
	rand.Seed(time.Now().UnixNano()) // initialize global pseudo random generator
	return Voices[rand.Intn(len(Voices))]
}

func (c *AzureClient) GetVoices(speakers map[string]struct{}) map[string]string {
	v := make(map[string]string)
	var i int
	for speaker := range speakers {
		v[speaker] = Voices[i]
		i++
	}
	return v
}

// download audio file from azure text-to-speech api if it doesn't exist in cache dir.
// we also store a sentenceAndDialogOnlyDir to create audio loops for which we want to exclude words and chars.
func (c *AzureClient) Fetch(ctx context.Context, query, filename string, isSentenceOrDialog bool) error {
	if contains(c.ignoreChars, query) {
		return nil
	}

	if err := os.MkdirAll(c.AudioDir, os.ModePerm); err != nil {
		return err
	}
	sentenceAndDialogOnlyDir := filepath.Join(c.AudioDir, "sentences_and_dialogs")
	if err := os.MkdirAll(sentenceAndDialogOnlyDir, os.ModePerm); err != nil {
		return err
	}
	lessonPath := filepath.Join(c.AudioDir, filename)
	cachePath := filepath.Join(c.AudioDir, "..", "..", "..", "audio", filename)
	sentenceAndDialogOnlyPath := filepath.Join(sentenceAndDialogOnlyDir, filename)

	// copy file from cache if exists, to lesson dir and to sentenceAndDialogOnlyDir
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
			return nil
		}
	}

	resp, err := c.fetch(ctx, query, 3)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(lessonPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	// for creating audio loops from sentences and dialogs only.
	if isSentenceOrDialog {
		if err := copyFileContents(lessonPath, sentenceAndDialogOnlyPath); err != nil {
			return err
		}
	}
	if err := copyFileContents(lessonPath, cachePath); err != nil {
		return err
	}

	if isSentenceOrDialog {
		fmt.Printf("audio content written to files:\n%s\n%s\n%s\n", lessonPath, cachePath, sentenceAndDialogOnlyPath)
	} else {
		fmt.Printf("audio content written to files:\n%s\n%s\n", lessonPath, cachePath)
	}
	return nil
}

func (c *AzureClient) fetch(ctx context.Context, text string, retryCount int) (*http.Response, error) {
	if retryCount == -1 {
		return nil, fmt.Errorf("excceded retries for query: %s", text)
	}

	query := `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="https://www.w3.org/2001/mstts" xml:lang="zh-CN">%s</speak>`
	query = fmt.Sprintf(query, text)

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer([]byte(query)))
	if err != nil {
		fmt.Printf("error creating request: %v", err)
		fmt.Println("retry...")
		return c.fetch(ctx, text, retryCount-1)
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", "audio-16khz-128kbitrate-mono-mp3")
	req.Header.Set("User-Agent", "curl")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error sending request to azure text-to-speech api: %v", err)
		fmt.Println("retry...")
		return c.fetch(ctx, query, retryCount-1)
	}
	return resp, nil
}

func (c *AzureClient) PrepareQueryWithRandomVoice(text string, addSplitAudio bool) string {
	return c.PrepareQuery(text, c.GetRandomVoice(), addSplitAudio)
}

// if text contains whitespaces and addSplitAudio is true, text is added twice, once with all
// whitespaces stipped off and once with whitespaces. azure api renders whitespaces as pauses in the audio.
func (c *AzureClient) PrepareQuery(text, speaker string, addSplitAudio bool) string {
	queryFmt := `
    <voice name="%s">
        <prosody rate="%s">
		    %s
        </prosody>
    </voice>`
	query := fmt.Sprintf(queryFmt, speaker, rate, strings.ReplaceAll(text, " ", ""))
	if addSplitAudio {
		query += fmt.Sprintf(queryFmt, speaker, rate, text)
	}

	return query
}
