from db import DataBaseHandler
from operator import itemgetter

dbh = DataBaseHandler()
data = dbh.get()
for media in data:
    id, media_type, title, year, overview, director, path = media
    print(
        f"{title}\t{title} ({year}){' -- '+director if director else ''}\t{overview}\t{path}"
    )
