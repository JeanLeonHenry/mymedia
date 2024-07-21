# Presentation
For now running `metadata.py` will
- lookup the media on [themoviedb.org](themoviedb.org), given a title and a release date (can be inferred from current directory name)
- download the poster if possible
- write poster and info on file and database

# Usage
Make a `.env` file so that the variables in `config.py` resolve properly.
Update config_path in `config.py` so that the config is loaded from `.env` file.

# TODO
Ideas to expand on
 - make a gui
 - forget about json, switch completely to db
