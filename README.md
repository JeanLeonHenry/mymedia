# Presentation
For now running `metadata.py` will
- lookup the media on [themoviedb.org](themoviedb.org), given a title and a release date (can be inferred from current directory name)
- download the poster if possible
- write poster and info on file and database

# Usage
Make a `.env` file so that the variables in `config.py` resolve properly.
Put that file in ~/.config/mymedia.
`picker.py` outputs some tab-separated info for easy parsing. See `picker.fish` for an example of quick and dirty TUI.

# TODO
Ideas to expand on
 - make a gui
