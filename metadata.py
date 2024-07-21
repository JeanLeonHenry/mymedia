#!/usr/bin/env python3
"""Download a movie/show info from themoviedb.org."""

import argparse
import os.path
import sys
from datetime import date
from io import BytesIO
from pathlib import Path
from typing import Any, Optional, Tuple

import requests
from pydantic import ValidationError

from config import (
    API_KEY,
    API_READ_TOKEN,
    API_URL,
    IMAGE_API_URL,
)
from db import DataBaseHandler
from models import Media, MediaCredits, MultiSearchResults, Person


def search_api(query: Optional[str] = None, endpoint: str = "search/multi") -> Any:
    """Given a query, search themoviedb.org API for results."""
    print(f"Searching TMDB.org on {endpoint} {'for '+query if query else ''}")

    headers = {
        "accept": "application/json",
        "Authorization": f"Bearer {API_READ_TOKEN}",
    }
    response = requests.get(
        API_URL + endpoint,
        headers=headers,
        params={
            "query": query,
        }
        if query
        else {},
    )
    response.raise_for_status()
    return response.json()


def parse_response(results, title: str, year: int, tolerance: int) -> Media:
    """Parse a requests response into a Media instance."""
    valid_response: MultiSearchResults = MultiSearchResults.model_validate(results)
    parsed_result: Optional[Media] = None
    for res in valid_response.results:
        # We don't care for persons : skip result if it's one
        try:
            Person.model_validate(res)
            continue
        except ValidationError:
            pass
        # It's not a person, validate it into a media
        try:
            parsed_result = Media.model_validate(res)
            if parsed_result.date and abs(parsed_result.date.year - year) <= tolerance:
                # accept this result
                break
            # validation went fine, but the year is wrong
            parsed_result = None
        except ValidationError:
            pass

    if parsed_result is None:
        sys.exit(
            f"Found no valid result for this query : {title} ({year}).\nIf you're looking for a foreign movie, especially with non latin alphabet, try using the original spelling.\nFirst result was : {valid_response.results[0]}"
        )
    return parsed_result


def get_poster(media: Media) -> Optional[bytes]:
    """Download a poster and save to disk."""
    print("Downloading poster.")
    if not media.poster_path:
        print("Tried to get the poster of a media without one")
        return
    url = IMAGE_API_URL + media.poster_path
    response = requests.get(url, params={"api_key": API_KEY})
    response.raise_for_status()
    img = BytesIO(response.content).read()
    return img


def get_director(movie: Media) -> None:
    """Look up movie director and update provided Media."""
    if movie.media_type != "movie":
        raise ValueError("Tried to get a director of media that isn't a movie.")
    results = search_api(endpoint=f"movie/{movie.id}/credits")
    valid_response = MediaCredits.model_validate(results)
    directors = [crew.name for crew in valid_response.crew if crew.job == "Director"]
    if len(directors) == 0:
        sys.exit("Found no director in movie credits")
    media.director = directors[0]


def parse_args() -> Tuple[str, int, argparse.Namespace]:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--info",
        nargs=2,
        help="Expects TITLE YEAR. If not provided, I will attempt to get that info from cwd name.",
    )
    parser.add_argument(
        "--tolerance",
        type=int,
        help="Tolerance for year lookup",
        default=2,
        choices=range(2, 6),
    )
    args = parser.parse_args()

    def year_format(y: str) -> int:
        try:
            res = date.fromisoformat(f"{y}-01-01").year
        except ValueError:
            sys.exit(f"Wrong input : year was {y}")
        return res

    title, year = None, None
    if args.info:
        title, year = args.info
        year = year_format(year)
    else:
        cwd_name = Path.cwd().stem.split(" (")
        if len(cwd_name) != 2:
            sys.exit(
                f"Wrong input : dir name was {Path.cwd().stem}. Should be 'TITLE (YEAR)'"
            )
        title, year = cwd_name
        year = year_format(year[:-1])
    return title, year, args


if __name__ == "__main__":
    title, year, args = parse_args()

    database = DataBaseHandler()
    if res := database.search(None, title, year):
        if len(res) == 7:
            id, media_type, title, year, overview, director, poster = res
            answer = input(
                f"Found {title} ({year}){' by '+director+' ' if director else ' '}in db.\nThe Movie DB page : https://www.themoviedb.org/{media_type}/{id}\nDo you want to replace local data? [y/N] "
            )
            if not answer.lower() == "y":
                sys.exit("Quitting.")

    results = search_api(title)
    media = parse_response(results, title, year, args.tolerance)
    if media.media_type == "movie":
        get_director(media)
    poster = get_poster(media)
    if not database.search(media.id):
        data = media.model_dump(exclude={"date", "poster_path"})
        data["path"] = os.path.join(os.getcwd())
        data_no_poster = data.copy()
        data["poster"] = poster
        database.push(data)
        print(f"Wrote {data_no_poster} media info to db (hiding poster).")
