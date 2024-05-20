package audio

// type Downloader struct {
// 	IgnoreChars []string
// 	AudioDir    string
// }

// // we support 4 different voices only
// var voices = []*texttospeechpb.VoiceSelectionParams{
// 	{
// 		LanguageCode: "cmn-CN",
// 		Name:         "cmn-CN-Wavenet-C",
// 		SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
// 	},
// 	{
// 		LanguageCode: "cmn-CN",
// 		Name:         "cmn-CN-Wavenet-A",
// 		SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
// 	},
// 	{
// 		LanguageCode: "cmn-CN",
// 		Name:         "cmn-TW-Wavenet-C",
// 		SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
// 	},
// 	{
// 		LanguageCode: "cmn-CN",
// 		Name:         "cmn-TW-Wavenet-A",
// 		SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
// 	},
// }

// func (p *Downloader) GetVoices(
// 	speakers map[string]struct{},
// ) map[string]*texttospeechpb.VoiceSelectionParams {
// 	v := make(map[string]*texttospeechpb.VoiceSelectionParams)
// 	var i int
// 	for speaker := range speakers {
// 		v[speaker] = voices[i]
// 		i++
// 	}
// 	return v
// }

// // download audio file from google text-to-speech api if it doesn't exist in cache dir.
// // we also store a sentenceAndDialogOnlyDir to create audio loops for which we want to exclude words and chars.
// func (p *Downloader) Fetch(ctx context.Context, query, filename string, isSentenceOrDialog bool) error {
// 	if contains(p.IgnoreChars, query) {
// 		return nil
// 	}
// 	if err := os.MkdirAll(p.AudioDir, os.ModePerm); err != nil {
// 		return err
// 	}
// 	sentenceAndDialogOnlyDir := filepath.Join(p.AudioDir, "sentences_and_dialogs")
// 	if err := os.MkdirAll(sentenceAndDialogOnlyDir, os.ModePerm); err != nil {
// 		return err
// 	}
// 	lessonPath := filepath.Join(p.AudioDir, filename)
// 	cachePath := filepath.Join(p.AudioDir, "..", "..", "..", "audio", filename)
// 	sentenceAndDialogOnlyPath := filepath.Join(sentenceAndDialogOnlyDir, filename)

// 	// copy file from cache to lesson dir and to sentenceAndDialogOnlyDir
// 	if _, err := os.Stat(cachePath); err == nil {
// 		var hasErr bool
// 		if err := copyFileContents(cachePath, lessonPath); err != nil {
// 			hasErr = true
// 			fmt.Printf("error copying cache file for query %s: %v\n", query, err)
// 		}
// 		if isSentenceOrDialog {
// 			if err := copyFileContents(cachePath, sentenceAndDialogOnlyPath); err != nil {
// 				hasErr = true
// 				fmt.Printf("error copying cache file for query %s: %v\n", query, err)
// 			}
// 		}
// 		if !hasErr {
// 			return nil
// 		}
// 	}

// 	resp, err := fetch(ctx, query, nil)
// 	if err != nil {
// 		return err
// 	}

// 	// the resp's AudioContent is binary.
// 	err = ioutil.WriteFile(lessonPath, resp.AudioContent, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}
// 	// for creating audio loops from sentences and dialogs only.
// 	if isSentenceOrDialog {
// 		err = ioutil.WriteFile(sentenceAndDialogOnlyPath, resp.AudioContent, os.ModePerm)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	err = ioutil.WriteFile(cachePath, resp.AudioContent, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Printf("%v\n", query)
// 	if isSentenceOrDialog {
// 		fmt.Printf("audio content written to files:\n%s\n%s\n%s\n", lessonPath, cachePath, sentenceAndDialogOnlyPath)
// 	} else {
// 		fmt.Printf("audio content written to files:\n%s\n%s\n", lessonPath, cachePath)
// 	}
// 	return nil
// }

// func (p *Downloader) FetchTmpAudioWithVoice(ctx context.Context, query, filename string, voice *texttospeechpb.VoiceSelectionParams) (string, error) {
// 	filename = filename + ".mp3"
// 	// cachePath := filepath.Join(p.AudioDir, "..", "..", "..", "audio", filename)
// 	tmpFile, err := os.CreateTemp("", "zh")
// 	if err != nil {
// 		return "", fmt.Errorf("could not create tmp file: %v", err)
// 	}

// 	// copy file from cache to tmp dir
// 	// if _, err := os.Stat(cachePath); err == nil {
// 	// 	var hasErr bool
// 	// 	if err := copyFileContents(cachePath, tmpFile.Name()); err != nil {
// 	// 		hasErr = true
// 	// 		fmt.Printf("error copying cache file for query %s: %v\n", query, err)
// 	// 	}
// 	// 	if !hasErr {
// 	// 		return tmpFile.Name(), nil
// 	// 	}
// 	// }

// 	resp, err := fetch(ctx, query, voice)
// 	if err != nil {
// 		return "", err
// 	}

// 	// The resp's AudioContent is binary.
// 	err = ioutil.WriteFile(tmpFile.Name(), resp.AudioContent, os.ModePerm)
// 	if err != nil {
// 		return "", err
// 	}

// 	fmt.Printf("%v\n", query)
// 	return tmpFile.Name(), nil
// }

// func (p *Downloader) JoinAndSaveDialogAudio(filename string, partsPaths []string) error {
// 	if err := os.MkdirAll(p.AudioDir, os.ModePerm); err != nil {
// 		return err
// 	}
// 	sentenceAndDialogOnlyDir := filepath.Join(p.AudioDir, "sentences_and_dialogs")
// 	if err := os.MkdirAll(sentenceAndDialogOnlyDir, os.ModePerm); err != nil {
// 		return err
// 	}
// 	lessonPath := filepath.Join(p.AudioDir, filename)
// 	cachePath := filepath.Join(p.AudioDir, "..", "..", "..", "audio", filename)
// 	sentenceAndDialogOnlyPath := filepath.Join(sentenceAndDialogOnlyDir, filename)

// 	if err := joinMP3Files(partsPaths, cachePath); err != nil {
// 		return err
// 	}
// 	if err := joinMP3Files(partsPaths, lessonPath); err != nil {
// 		return err
// 	}
// 	if err := joinMP3Files(partsPaths, sentenceAndDialogOnlyPath); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func joinMP3Files(inputPaths []string, outputPath string) error {
// 	// Generate a FFmpeg command to join the MP3 files
// 	ffmpegArgs := []string{"-i", "concat:" + strings.Join(inputPaths, "|"), "-c", "copy", "-y", outputPath}

// 	// Execute the FFmpeg command
// 	cmd := exec.Command("ffmpeg", ffmpegArgs...)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	if err := cmd.Run(); err != nil {
// 		return fmt.Errorf("failed to join MP3 files: %v", err)
// 	}

// 	fmt.Printf("audio content written to files:\n%s\n", outputPath)
// 	return nil
// }

// // uses a random voice if param voice is nil
// func fetch(ctx context.Context, query string, voice *texttospeechpb.VoiceSelectionParams) (*texttospeechpb.SynthesizeSpeechResponse, error) {
// 	time.Sleep(100 * time.Millisecond)
// 	client, err := texttospeech.NewClient(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer client.Close()

// 	if voice == nil {
// 		rand.Seed(time.Now().UnixNano()) // initialize global pseudo random generator
// 		voice = voices[rand.Intn(len(voices))]
// 	}
// 	// Perform the text-to-speech request on the text input with the selected
// 	// voice parameters and audio file type.
// 	req := texttospeechpb.SynthesizeSpeechRequest{
// 		// Set the text input to be synthesized.
// 		Input: &texttospeechpb.SynthesisInput{
// 			InputSource: &texttospeechpb.SynthesisInput_Text{Text: query},
// 		},
// 		// Build the voice request, select the language code ("en-US") and the SSML
// 		// voice gender ("neutral").
// 		Voice: voice,
// 		// Select the type of audio file you want returned.
// 		AudioConfig: &texttospeechpb.AudioConfig{
// 			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
// 			SpeakingRate:  0.85,
// 		},
// 	}
// 	return client.SynthesizeSpeech(ctx, &req)
// }

// func contains[T comparable](s []T, e T) bool {
// 	for _, v := range s {
// 		if v == e {
// 			return true
// 		}
// 	}
// 	return false
// }

// func copyFileContents(src, dst string) (err error) {
// 	in, err := os.Open(src)
// 	if err != nil {
// 		return
// 	}
// 	defer in.Close()
// 	out, err := os.Create(dst)
// 	if err != nil {
// 		return
// 	}
// 	defer func() {
// 		cerr := out.Close()
// 		if err == nil {
// 			err = cerr
// 		}
// 	}()
// 	if _, err = io.Copy(out, in); err != nil {
// 		return
// 	}
// 	err = out.Sync()
// 	return
// }
