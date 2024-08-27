#!/usr/bin/env bash

folder="$("$HOME"/go/bin/mymedia picker)"
echo "Got folder: $folder"
if [[ -n "$folder" ]]; then
	notify-send "$(basename "$folder")"
	swallow mpv "$folder"
else
	notify-send "No media"
fi
