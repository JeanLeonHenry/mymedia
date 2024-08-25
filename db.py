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
        """Write data to database.

        Args:
            data (dict):

        """
        cur = self.connection.cursor()
        cur.execute(
            "INSERT INTO media VALUES(:id, :media_type, :title, :year, :overview, :director, :poster, :path)",
            data,
        )
        self.connection.commit()

    # NOTE: useless ?
    def get(self, size: int | None, posters: bool = False) -> List[tuple]:
        """Get all, or some, of the rows in the db.

        Args:
            size (int):
            posters (bool):

        Returns:
            (list): results


        """
        cur = self.connection.cursor()
        query = f"SELECT id, media_type, title, year, overview, director{', poster' if posters else ''}, path FROM media ORDER BY title, year ASC"
        if size is not None and size > 0:
            res = cur.execute(query).fetchmany(size)
        else:
            res = cur.execute(query).fetchall()
        return res

    def getByInfo(self, title: str, year: int):
        cur = self.connection.cursor()
        query = "SELECT id, media_type, title, year, overview, director, poster, path FROM media WHERE media.title=? AND media.year=?"
        res = cur.execute(query, (title, year)).fetchall()
        if len(res) > 1:
            raise ValueError("Found several media, try by id")
        if len(res) == 0:
            raise ValueError("Found no media.")
        return res[0]

    def searchById(self, id: int) -> Optional[Tuple[str, int]]:
        """Query the db by TMDB id.

        Args:
            id (int):

        Returns:
            a tuple containing the title and the year

        """
        cur = self.connection.cursor()
        query = "SELECT title, year FROM media WHERE media.id = ?"
        data = (id,)
        # FIX: handle the case where there's several results
        res = cur.execute(query, data).fetchone()
        return res

    def searchByInfo(self, title: str, year: int) -> Optional[Tuple[int]]:
        """Query the db by title and release year.

        Args:
            title (str):
            year (int):

        Returns:
            a tuple with an integer TMDB id


        """
        cur = self.connection.cursor()
        query = "SELECT id FROM media WHERE media.title = ? AND media.year = ?"
        data = (title, year)
        # FIX: handle the case where there's several results
        res = cur.execute(query, data).fetchone()
        return res

    # NOTE: useless ?
    def update_path(self, id: int, data: str) -> None:
        cur = self.connection.cursor()
        cur.execute("UPDATE media SET path=? WHERE media.id=?", (data, id))
        self.connection.commit()
