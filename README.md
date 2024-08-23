# Presentation

## Metadata lookup
For now, running `metadata.py` will
- lookup the media on [themoviedb.org](themoviedb.org), given a title and a release year. That info can be
    - provided with the `--info` option flag
    - inferred from current directory name (must be in the format `TITLE (YEAR)`, parentheses included)
- write data to a database, poster included

## Media picker
- `myMediaUI/myMediaUI` lets you choose the media you want to play and output the media directory to STDIN. Then use your favorite media player. For example, see `picker.sh`

# Usage
- Make a `.env` file so that the variables in `config.py` resolve properly.
- Put that file in `~/.config/mymedia`.

# TODO
 - The current picker should grab the poster image from the db.
