data_dir=./data/$(source)
audio_dir=./data/$(source)/audio

JSON_CACHE=/home/f/Dropbox/zh/cache/cards
AUDIO_CACHE=/home/f/Dropbox/zh/cache/audio

.PHONY: clean
clean:
	rm -r $(data_dir)/output || true
	rm -r $(audio_dir) || true

.PHONY: gen
gen: clean
	go run cmd/main.go -src $(source)

anki_audio_dir="/home/f/.local/share/Anki2/User 1/collection.media/"
.PHONY: cp-audio
cp-audio:
	@cd $(audio_dir)
	@echo "copy audio files to anki audio dir: $(anki_audio_dir)"
	$(shell find $(audio_dir) -type f -name '*.mp3' -exec cp {} $(anki_audio_dir) \;)
	@echo "copy audio files to cache: $(AUDIO_CACHE)"
	$(shell cp $(audio_dir)/* $(AUDIO_CACHE))

.PHONY: anki
anki: gen cp-audio
	@echo "don't forget to commit ignore file!"

.PHONY: anki-dry
anki-dry: clean
	go run cmd/main.go -src $(source) -dryrun

.PHONY: cp-json
cp-json:
	mkdir -p $(JSON_CACHE)/words $(JSON_CACHE)/sentences $(JSON_CACHE)/clozes $(JSON_CACHE)/grammar
	cp $(data_dir)/output/words/* $(JSON_CACHE)/words || true
	cp $(data_dir)/output/sentences/* $(JSON_CACHE)/sentences || true
	cp $(data_dir)/output/clozes/* $(JSON_CACHE)/clozes || true
	cp $(data_dir)/output/grammar/* $(JSON_CACHE)/grammar || true

.PHONY: fetch-daily
fetch-daily:
	go run cmd/anki-connect/fetch-due/main.go

.PHONY: segment
segment:
	@rm /tmp/segmented || true
	@if [ -f $(data_dir)/dialogues ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/dialogues UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/dialogues;fi
	@if [ -f $(data_dir)/sentences ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/sentences UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/sentences;fi
	@if [ -f $(data_dir)/clozes ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/clozes UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/clozes;fi
	@if [ -f $(data_dir)/patterns ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/patterns UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/patterns;fi

.PHONY: lint
lint:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Running golangci-lint..."
	@golangci-lint run ./...
