data_dir=./data/$(source)/$(lesson)
audio_dir=./data/$(source)/$(lesson)/audio

.PHONY: gen
gen:
	rm -r -f $(data_dir)/output/
	mkdir $(data_dir)/output/
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

.PHONY: segment
segment:
	@rm /tmp/segmented || true
	@if [ -f $(data_dir)/input/dialogues ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/dialogues UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/dialogues;fi
	@if [ -f $(data_dir)/input/sentences ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/sentences UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/sentences;fi
	@if [ -f $(data_dir)/input/clozes ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/clozes UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/clozes;fi
