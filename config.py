import os
from dotenv import dotenv_values

API_URL = "https://api.themoviedb.org/3/"
IMAGE_API_URL = "http://image.tmdb.org/t/p/w500"

config_path = os.path.join(os.getenv("HOME"), ".config/mymedia/.env")
config = dotenv_values(config_path)
API_KEY = config["API_KEY"]
API_READ_TOKEN = config["API_READ_TOKEN"]
DB_PATH = config["DB_PATH"]
