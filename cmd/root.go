package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JeanLeonHenry/mymedia/config"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "mymedia",
	Short:   "Build and query a media library.",
	Long:    ``,
	Version: time.Now().Format(time.DateTime),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}
var localConfig *config.Config
var debug bool

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	localConfig = config.New()
	localConfig.Check()
	msg := fmt.Sprintf("Was config valid ? %v\nConfig was : %+v", localConfig.IsValid, localConfig)
	if !localConfig.IsValid {
		log.Fatal(msg)
	}
	// NOTE: use viper for better config handling ?
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "add extra logging")
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mymedia.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
