from dotenv import dotenv_values

poster_filename = "poster.png"
info_filename = "media_info.json"
config_path = ""
API_URL = "https://api.themoviedb.org/3/"
IMAGE_API_URL = "http://image.tmdb.org/t/p/w500"

config = dotenv_values(config_path)
API_KEY = config["API_KEY"]
API_READ_TOKEN = config["API_READ_TOKEN"]
DB_PATH = config["DB_PATH"]
