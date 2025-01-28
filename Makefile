data_dir=./data/$(source)/$(lesson)
audio_dir=./data/$(source)/$(lesson)/audio

JSON_CACHE=/home/f/Dropbox/data/zh/cache/json

.PHONY: gen
gen:
	mkdir -p $(audio_dir) || true
	go run cmd/main.go -l $(lesson) -src $(source) -tgt $(target) -t $(tags)

anki_audio_dir="/home/f/.local/share/Anki2/User 1/collection.media/"
.PHONY: cp-audio
cp-audio:
	@cd $(audio_dir)
	@echo "copy audio files to anki audio dir: $(anki_audio_dir)"
	$(shell find $(audio_dir) -type f -name '*.mp3' -exec cp {} $(anki_audio_dir) \;)

.PHONY: anki
anki: gen cp-audio
	@echo "don't forget to commit ignore file!"

.PHONY: anki-dry
anki-dry:
	go run cmd/main.go -l $(lesson) -src $(source) -dryrun
	mkdir -p $(JSON_CACHE)/words $(JSON_CACHE)/sentences
	cp $(data_dir)/output/words/* $(JSON_CACHE)/words || true
	cp $(data_dir)/output/sentences/* $(JSON_CACHE)/sentences || true

.PHONY: segment
segment:
	@rm /tmp/segmented || true
	@if [ -f $(data_dir)/input/dialogues ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/dialogues UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/dialogues;fi
	@if [ -f $(data_dir)/input/sentences ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/sentences UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/sentences;fi
	@if [ -f $(data_dir)/input/clozes ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/clozes UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/clozes;fi

.PHONY: lint
lint:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Running golangci-lint..."
	@golangci-lint run ./...
