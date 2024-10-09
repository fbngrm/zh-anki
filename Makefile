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

.PHONY: new
new:
	mkdir -p $(data_dir)/input
	touch  $(data_dir)/input/dialogues
	touch  $(data_dir)/input/sentences
	touch  $(data_dir)/input/words
	touch  $(data_dir)/audio

.PHONY: audio_concat
silence=../../../../silence_64kb.mp3
audio_concat_dir=./data/$(source)/$(lesson)/audio/sentences_and_dialogs
out_dir=./data/$(source)/$(lesson)/audio/sentences_and_dialogs/concat
audio_concat:
	mkdir -p $(out_dir)
	cd $(audio_concat_dir); for i in *.mp3; do ffmpeg -i "$$i" -filter:a "atempo=0.7" /tmp/"$${i%.*}_slow.mp3"; done
	cd $(audio_concat_dir); for i in *.mp3; do ffmpeg -i "concat:$$i|$(silence)|/tmp/$${i%.*}_slow.mp3|$(silence)|$$i|$(silence)|/tmp/$${i%.*}_slow.mp3|$(silence)|$$i|$(silence)|$(silence)" -acodec copy ./concat/"$${i%.*}_concat.mp3"; done

.PHONY: mv-audio
src=./data/$(source)/$(lesson)/audio/sentences_and_dialogs/concat
dst=/home/f/data/rslsync/zh/$(source)/$(lesson)
mv-audio:
	mkdir -p $(dst)
	echo "move files to $(dst)"
	cp $(src)/*concat.mp3 $(dst)/

.PHONY: audio
audio_concat_dir=./data/$(source)/$(lesson)/audio/sentences_and_dialogs
out_dir=./data/$(source)/$(lesson)/audio/sentences_and_dialogs/concat
audio:
	mkdir -p $(out_dir)
	cd $(audio_concat_dir); for i in *.mp3; do ffmpeg -i "$$i" -filter:a "atempo=0.7" /tmp/"$${i%.*}_slowed.mp3"; done
	cd $(audio_concat_dir); for i in *.mp3; do ffmpeg -i  /tmp/"$${i%.*}_slowed.mp3" -af "apad=pad_dur=1"  /tmp/"$${i%.*}_slowed_silence.mp3"; done
	cd $(audio_concat_dir); for i in *.mp3; do ffmpeg -i  "$$i" -af "apad=pad_dur=1"  /tmp/"$${i%.*}_silence.mp3"; done
	cd $(audio_concat_dir); for i in *.mp3; do ffmpeg -i /tmp/"$${i%.*}_silence.mp3" -i /tmp/"$${i%.*}_slowed_silence.mp3" -filter_complex "[0:a][1:a][0:a][1:a][0:a][1:a]concat=n=6:v=0:a=1[out]" -map "[out]" ./concat/"$${i%.*}_concat.mp3"; done

.PHONY: segment
segment:
	@rm /tmp/segmented || true
	@if [ -f $(data_dir)/input/dialogues ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/dialogues UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/dialogues;fi
	@if [ -f $(data_dir)/input/sentences ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/sentences UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/sentences;fi
	@if [ -f $(data_dir)/input/clozes ]; then cd ../stanford-segmenter && ./segment.sh pku ../zh-anki/$(data_dir)/input/clozes UTF-8 0 > /tmp/segmented && cat /tmp/segmented > ../zh-anki/$(data_dir)/input/clozes;fi
