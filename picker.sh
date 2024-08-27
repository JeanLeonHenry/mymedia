#!/usr/bin/env bash

folder="$(mymedia picker)"
echo "Got folder: $folder"
if [[ -n "$folder" ]]; then
	notify-send "$(basename "$folder")"
	swallow mpv "$folder"
else
	notify-send "No media"
fi
