package config

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/profclems/go-dotenv"
)

type Config struct {
	DefaultTolerance int
	// ApiUrl           string
	// ImageApiUrl      string
	ApiReadToken string
	ApiKey       string
	IsValid      bool
	Path         string
}

func New() *Config {
	dotenv.SetConfigFile(path.Join(os.Getenv("HOME"), ".config/mymedia/.env"))
	dbPath := dotenv.GetString("DB_PATH")
	if dbPath == "" {
		log.Fatal("DB_PATH is empty, check config file.")
	}

	return &Config{
		DefaultTolerance: 2,
		IsValid:          true,
		Path:             dbPath,
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
