package cmd

import (
	"log"
	"os"
	"path"

	"github.com/JeanLeonHenry/mymedia/internal/db"
	"github.com/profclems/go-dotenv"
	"github.com/spf13/cobra"
)

// HACK: global var for config
var config struct {
	DBH              *db.DBHandler
	DefaultTolerance int
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mymedia",
	Short: "Build and query a media library.",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// NOTE: use viper for better config handling ?
	dotenv.SetConfigFile(path.Join(os.Getenv("HOME"), ".config/mymedia/.env"))
	dbPath := dotenv.GetString("DB_PATH")
	if dbPath == "" {
		log.Fatal("db path is empty, check config file.")
	}
	config.DBH = db.NewDB(dbPath)
	config.DefaultTolerance = 2
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mymedia.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
