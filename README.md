# Usage
Build and query a media library.

```
Usage:
  mymedia [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  picker      TUI to query the database
  poster      Given a title, reads poster from db and write it in cwd
  scan        Scans the current folder for media folders and update database

Flags:
  -d, --debug   add extra logging
  -h, --help    help for mymedia

Use "mymedia [command] --help" for more information about a command.
```

## Poster
the config (see below) must provide a path to `.db` file that contains a table created with
```sql
CREATE TABLE media(id, media_type, title, year, overview, director, poster, path)
```
the `poster` field holds the raw bytes for the poster image downloaded from TMDB.

# Configuration
- Make a `.env` file so that the variables in `config/config.go` resolve properly.
- Put that file in `~/.config/mymedia`.

# TODO
 - The current picker should grab the poster image from the db.
