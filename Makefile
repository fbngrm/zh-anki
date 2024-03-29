data_dir=./data/$(source)/$(lesson)
audio_dir=./data/$(source)/$(lesson)/audio

.PHONY: gen
gen:
	rm -r -f $(data_dir)/output/
	mkdir $(data_dir)/output/
	rm -r -f $(audio_dir)
	mkdir $(audio_dir)
	go run cmd/main.go -l $(lesson) -src $(source) -tgt $(target) -t $(tags)

anki_audio_dir="/home/f/.local/share/Anki2/User 1/collection.media/"
.PHONY: cp-audio
cp-audio:
	cd $(audio_dir)
	$(shell find $(audio_dir) -type f -name '*.mp3' -exec cp {} $(anki_audio_dir) \;)

.PHONY: anki
anki: gen cp-audio
	@echo "don't forget to commit ignore file!"

.PHONY: new
new:
	mkdir -p $(data_dir)
	touch  $(data_dir)/dialog
	touch  $(data_dir)/sentences
	touch  $(data_dir)/words

# .PHONY: commit-ignore
# ignore_path=$(data_dir)/../ignore
# prev_ignore_path=./data/prev_ignore_commit
# commit-ignore:
# 	$(shell git add $(ignore_path))
# 	$(shell git commit -m "commit ignore for lesson $(lesson)")
# 	$(shell git rev-parse HEAD > $(prev_ignore_path))

# .PHONY: reset-ignore
# reset-ignore:
# 	@echo $(prev_ignore_path)
# 	$(shell git revert $(shell cat $(prev_ignore_path)))
# 	rm $(prev_ignore_path)

# .PHONY: reset-files
# reset-files:
# 	rm $(data_dir)/cards.md $(data_dir)/dialog*

# .PHONY: reset
# reset: reset-ignore reset-files

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
