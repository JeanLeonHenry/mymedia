#!/usr/bin/env bash

folder=$( ~/prog/github.com/JeanLeonHenry/mymedia/myMediaUI/myMediaUI )
notify-send "$(basename "$folder")"
swallow mpv "$folder"
