# Presentation
For now running `metadata.py` will
- lookup the media on [themoviedb.org](themoviedb.org), given a title and a release date (can be inferred from current directory name)
- write data to a database, poster include

# Usage
- Make a `.env` file so that the variables in `config.py` resolve properly.
- Put that file in `~/.config/mymedia`.
- `picker.py` outputs some tab-separated info for easy parsing. See `picker.fish` for an example of quick and dirty TUI.

# TODO
 - make a proper GUI, that takes the poster image from the db.
