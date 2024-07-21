import sqlite3
from typing import Optional

from config import DB_PATH


class DataBaseHandler:
    def __init__(self) -> None:
        self.connection = sqlite3.connect(DB_PATH)
        # If no table is found, create it
        cur = self.connection.cursor()
        res = cur.execute("SELECT name FROM sqlite_master")
        if len(res.fetchall()) == 0:
            cur.execute(
                "CREATE TABLE media(id, media_type, title, year, overview, director, poster)"
            )
            self.connection.commit()

    def push(self, data: dict):
        cur = self.connection.cursor()
        cur.execute(
            "INSERT INTO media VALUES(:id, :media_type, :title, :year, :overview, :director, :poster)",
            data,
        )
        self.connection.commit()

    def search(
        self,
        id: Optional[int] = None,
        title: Optional[str] = None,
        year: Optional[int] = None,
    ) -> tuple:
        """Query the db by either an id or (title, year)."""
        cur = self.connection.cursor()
        query = None
        data = None
        if id is not None:
            query = "SELECT title, year FROM media WHERE media.id = ?"
            data = (id,)
        elif title and year:
            query = "SELECT id FROM media WHERE media.title = ? AND media.year = ?"
            data = (title, year)
        else:
            raise ValueError("Wrong arguments")
        # FIX: handle the case where there's several results
        res = cur.execute(query, data).fetchone()
        if res is None:
            return ()
        return res
