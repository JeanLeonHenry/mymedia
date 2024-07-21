from dotenv import dotenv_values

poster_filename = "poster.png"
info_filename = "media_info.json"
config_path = ""
config = dotenv_values(config_path)
API_KEY = config["API_KEY"]
API_READ_TOKEN = config["API_READ_TOKEN"]
DB_PATH = config["DB_PATH"]
API_URL = config["API_URL"]
IMAGE_API_URL = config["IMAGE_API_URL"]
