import sqlite3
from typing import List, Optional, Tuple

from config import DB_PATH


class DataBaseHandler:
    def __init__(self) -> None:
        self.connection = sqlite3.connect(DB_PATH)
        # If no table is found, create it
        cur = self.connection.cursor()
        res = cur.execute("SELECT name FROM sqlite_master")
        if len(res.fetchall()) == 0:
            cur.execute(
                "CREATE TABLE media(id, media_type, title, year, overview, director, poster, path)"
            )
            self.connection.commit()

    def write(self, data: dict):
        cur = self.connection.cursor()
        cur.execute(
            "INSERT INTO media VALUES(:id, :media_type, :title, :year, :overview, :director, :poster, :path)",
            data,
        )
        self.connection.commit()

    def get(self, size: int = -1, posters: bool = False) -> List[tuple]:
        cur = self.connection.cursor()
        query = f"SELECT id, media_type, title, year, overview, director{', poster' if posters else ''}, path FROM media ORDER BY title, year ASC"
        if size != -1:
            res = cur.execute(query).fetchmany(size)
        else:
            res = cur.execute(query).fetchall()
        return res

    def searchById(self, id: int) -> Optional[Tuple[str, int]]:
        cur = self.connection.cursor()
        query = "SELECT title, year FROM media WHERE media.id = ?"
        data = (id,)
        # FIX: handle the case where there's several results
        res = cur.execute(query, data).fetchone()
        return res

    def searchByInfo(self, title: str, year: int) -> Optional[Tuple[int]]:
        cur = self.connection.cursor()
        query = "SELECT id FROM media WHERE media.title = ? AND media.year = ?"
        data = (title, year)
        # FIX: handle the case where there's several results
        res = cur.execute(query, data).fetchone()
        return res

    def update_path(self, id: int, data: str) -> None:
        cur = self.connection.cursor()
        cur.execute("UPDATE media SET path=? WHERE media.id=?", (data, id))
        self.connection.commit()
