from pydantic import BaseModel, AliasChoices, Field, computed_field
from typing import List, Literal, Optional
from datetime import datetime


class Person(BaseModel):
    media_type: Literal["person"]


class Media(BaseModel):
    id: int
    media_type: Literal["movie", "tv"]
    title: str = Field(validation_alias=AliasChoices("name", "title"))
    date: datetime = Field(
        validation_alias=AliasChoices("release_date", "first_air_date")
    )
    overview: str
    poster_path: Optional[str] = Field(
        pattern=r"^/[a-zA-Z0-9]+\.((jpg)|(jpeg)|(png)|(gif)|(bmp))$"
    )
    director: str = Field(default="")

    @computed_field
    def year(self) -> int:
        return self.date.year


Result = Media | Person


class MultiSearchResults(BaseModel):
    total_results: int = Field(ge=1)
    results: List


class Crew(BaseModel):
    job: str
    name: str


class MediaCredits(BaseModel):
    crew: List[Crew]
