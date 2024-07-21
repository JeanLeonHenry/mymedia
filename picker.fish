function mymedia
    set folder (python3 ~/prog/github.com/JeanLeonHenry/mymedia/picker.py|fzf -d "\t" --with-nth=1 --preview='
    echo {2}
    echo
    echo {3}|fold -w ${FZF_PREVIEW_COLUMNS} -s 
    COLS=$((LINES*2/3))
    kitten icat --clear --transfer-mode=memory --stdin=no --unicode-placeholder --place=${COLS}x${FZF_PREVIEW_LINES}@0x0 {-1}/poster.*'|cut -f 4)
    cd $folder
    if test $HOME != (pwd)
        notify-send (basename (pwd))
        fd -tf "mp4|mkv|avi|mpg\$" -X swallow mpv $folder
    end
end
mymedia
