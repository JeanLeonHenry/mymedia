package config

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/JeanLeonHenry/mymedia/internal/db"
	"github.com/profclems/go-dotenv"
)

type Config struct {
	DBH              *db.DBHandler
	DefaultTolerance int
	// ApiUrl           string
	// ImageApiUrl      string
	ApiReadToken string
	ApiKey       string
	IsValid      bool
}

func New() *Config {
	dotenv.SetConfigFile(path.Join(os.Getenv("HOME"), ".config/mymedia/.env"))
	dbPath := dotenv.GetString("DB_PATH")
	if dbPath == "" {
		log.Fatal("DB_PATH is empty, check config file.")
	}

	return &Config{
		DBH:              db.NewDB(dbPath),
		DefaultTolerance: 2,
		IsValid:          true,
	}

}

func (c *Config) warningStringVarEmpty(val string, key string) {
	if val == "" {
		log.Printf("%v is empty, check config file.", key)
		c.IsValid = false
	}
}

func (c *Config) Check() {
	configKeys := map[string]*string{
		// "API_URL":        &c.ApiUrl,
		// "IMAGE_API_URL":  &c.ImageApiUrl,
		"API_READ_TOKEN": &c.ApiReadToken,
		"API_KEY":        &c.ApiKey,
	}
	for key, configField := range configKeys {
		*configField = dotenv.GetString(key)
		c.warningStringVarEmpty(*configField, key)
	}
}

func (c Config) String() string {
	out, _ := json.MarshalIndent(c, "", "	")
	return string(out)
}
