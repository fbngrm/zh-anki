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

	"golang.org/x/exp/slog"
	"golang.org/x/net/http2"
)

// speed
const rate = "0.7"
const maxRetries = 10

type AzureClient struct {
	cache       *Cache
	endpoint    string
	apiKey      string
	AudioDir    string
	ignoreChars []string
}

func NewAzureClient(endpoint, apiKey, audioDir string, ignoreChars []string, cache *Cache) *AzureClient {
	return &AzureClient{
		endpoint:    endpoint,
		apiKey:      apiKey,
		AudioDir:    audioDir,
		ignoreChars: ignoreChars,
		cache:       cache,
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
	filename = strings.ReplaceAll(filename, " ", "")
	if c.cache.Get(filename) {
		slog.Debug("fetch azure audio, query found in cache")
		return nil
	}
	// time.Sleep(2000 * time.Millisecond)
	if retryCount <= 0 {
		slog.Error("download azure audio", "error", "exceeded retries", "query", query)
		return nil
	}
	if contains(c.ignoreChars, query) {
		return nil
	}

	if err := os.MkdirAll(c.AudioDir, os.ModePerm); err != nil {
		return err
	}
	lessonPath := filepath.Join(c.AudioDir, filename)

	resp, err := c.fetch(ctx, query, maxRetries)
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

	slog.Debug("download audio", "path", lessonPath)
	return nil
}

func (c *AzureClient) fetch(ctx context.Context, query string, retryCount int) (*http.Response, error) {
	if retryCount == -1 {
		return nil, fmt.Errorf("excceded retries for query: %s", query)
	}
	if retryCount == maxRetries {
		query = fmt.Sprintf(`<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="https://www.w3.org/2001/mstts" xml:lang="zh-CN">%s</speak>`, query)
	}
	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer([]byte(query)))
	if err != nil {
		fmt.Printf("error creating request: %v", err)
		fmt.Println("retry...")
		return c.fetch(ctx, query, retryCount-1)
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

	if resp.StatusCode != http.StatusOK {
		buf := new(strings.Builder)
		_, err = io.Copy(buf, resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		s := buf.String()
		if s == "Quota Exceeded" || resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(5000 * time.Millisecond)
			return c.fetch(ctx, query, retryCount-1)
		}
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
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

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
