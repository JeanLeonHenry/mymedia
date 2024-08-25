# Usage
Build and query a media library.
```
Usage:
  mymedia [command]

Available Commands:
  picker      TUI to query the database
  poster      Reads poster from db and write it in cwd

WIP:
  scan        Scans the current folder for media folders and update database
```
## Picker
Provides a fzf-based TUI to query the database.
The output will be the path to the selected media directory.
⚠️ external dependencies: fold, kitty

## Poster
the config (see below) must provide a path to `.db` file that is the result of the following sqlite statement
```sql
CREATE TABLE media(id, media_type, title, year, overview, director, poster, path)
```
the `poster` field holds the raw bytes for the poster image downloaded from TMDB.

```
Usage:
  mymedia poster [flags]

Flags:
  -r, --replace        if replace is true, replace file if it exists
  -t, --title string   media title, case insensitive, will be read from cwd name if missing
```
# Configuration
- Make a `.env` file so that the variables in `config.py` resolve properly.
- Put that file in `~/.config/mymedia`.

# TODO
 - The current picker should grab the poster image from the db.
 - implement the scan command in go
