package audio

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/go-mp3"

	"golang.org/x/exp/slog"
	"golang.org/x/net/http2"
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
	"zh-CN-XiaoxiaoNeural", // female
	"zh-CN-YunjianNeural",  // male
	"zh-CN-XiaochenNeural", // female
	// "zh-CN-YinyangNeural",  // male / broken
	"zh-CN-YunyiMultilingualNeural", // male
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
func (c *AzureClient) Fetch(ctx context.Context, query, filename string, retryCount int) error {
	time.Sleep(500 * time.Millisecond)
	if retryCount <= 0 {
		slog.Error("download azure audio", "error", "excceded retries", "query", query)
		return nil
	}
	if contains(c.ignoreChars, query) {
		return nil
	}

	if err := os.MkdirAll(c.AudioDir, os.ModePerm); err != nil {
		return err
	}
	lessonPath := filepath.Join(c.AudioDir, filename)
	cachePath := filepath.Join(c.AudioDir, "..", "..", "..", "audio", filename)

	// copy file from cache if exists, to lesson dir and to sentenceAndDialogOnlyDir
	if _, err := os.Stat(cachePath); err == nil {
		var hasErr bool
		if err := copyFileContents(cachePath, lessonPath); err != nil {
			hasErr = true
			slog.Error("copy cache file", "query", query, "error", err)
		}
		if !hasErr {
			slog.Debug("copy cache file", "query", query)
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

	if err := checkMP3Integrity(lessonPath); err != nil {
		slog.Error("download audio file", "error", err)
		return c.Fetch(ctx, query, filename, retryCount-1)
	}

	slog.Debug("download audio", "path", lessonPath)
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

	client := &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error sending request to azure text-to-speech api: %v", err)
		fmt.Println("retry...")
		return c.fetch(ctx, query, retryCount-1)
	}
	return resp, nil
}

func (c *AzureClient) PrepareQueryWithRandomVoice(text string, addSplitAudio bool) string {
	speaker := c.GetRandomVoice()
	return c.PrepareQuery(text, speaker, addSplitAudio)
}

// if text contains whitespaces and addSplitAudio is true, text is added twice, once with all
// whitespaces stipped off and once with whitespaces. azure api renders whitespaces as pauses in the audio.
func (c *AzureClient) PrepareQuery(text, speaker string, addSplitAudio bool) string {
	slog.Debug("prepare azure query", "voice", speaker, "text", text)
	queryFmt := `<voice name="%s"><prosody rate="%s">%s</prosody></voice>`
	query := fmt.Sprintf(queryFmt, speaker, rate, strings.ReplaceAll(text, " ", ""))
	if addSplitAudio {
		query += fmt.Sprintf(queryFmt, speaker, rate, text)
	}
	return query
}

// checkMP3Integrity tries to decode an MP3 file and checks for errors.
func checkMP3Integrity(filename string) error {
	// Open the MP3 file
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Attempt to decode the MP3 file
	_, err = mp3.NewDecoder(f)
	if err != nil {
		return fmt.Errorf("file is not a valid MP3: %w", err)
	}

	// If no error, the MP3 file is valid
	return nil
}
