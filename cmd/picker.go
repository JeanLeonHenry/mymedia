package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	fzf "github.com/junegunn/fzf/src"

	_ "modernc.org/sqlite"
)

var pickerCmd = &cobra.Command{
	Use:   "picker",
	Short: "TUI to query the database",
	Long: `Provides a fzf-based TUI to query the database.
The output will be the path to the selected media directory.
External dependencies: fold, kitty
`,
	Run: func(cmd *cobra.Command, args []string) {

		inputChan := make(chan string)
		go func() {
			results, err := queries.ListMedia(ctx)
			if err != nil {
				log.Fatal("Query error : ", err)
			}
			for _, result := range results {
				if result.Director.Valid && result.Director.String != "" {
					result.Director.String = " — " + result.Director.String
				}
				infoLine := fmt.Sprintf("%v (%v)%v", result.Title, result.Year, result.Director.String)
				// TODO: use that folder mod time to influence sorting, see fzf docs
				fileStat, err := os.Stat(result.Path) // WARN: missing records are ignored
				var overview string
				if !result.Overview.Valid {
					overview = "No overview available."
				} else {
					overview = result.Overview.String
				}
				if err != nil {
					s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", result.Title, infoLine, overview, result.ID, result.Path, 0)
					inputChan <- s
				} else {
					s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%s", result.Title, infoLine, overview, result.ID, result.Path, fileStat.ModTime())
					inputChan <- s
				}
			}
			close(inputChan)
		}()

		exit := func(code int, err error) {
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			os.Exit(code)
		}

		cmdLineOptions := []string{"--delimiter=\\t", "--with-nth=1", "--accept-nth=-2"}

		// Use sqlite3 cli extension to temporarily write poster image data to disk
		posterFilePath := "/tmp/mymedia_poster.jpg"
		query := fmt.Sprintf(`SELECT writefile("%v", poster) FROM media WHERE id={-3}`, posterFilePath)
		// On focus of a line, execute the above sqlite query
		cmdLineOptions = append(cmdLineOptions, `--bind=focus:execute-silent(sqlite3 `+localConfig.Path+` '`+query+`')`)
		// Preview window setup
		previewCmd := "echo {2};echo;echo {3}|fold -w ${FZF_PREVIEW_COLUMNS} -s;COLS=$((LINES*2/3));kitten icat --clear --transfer-mode=memory --stdin=no --unicode-placeholder --place=${COLS}x${FZF_PREVIEW_LINES}@0x0 " + posterFilePath
		cmdLineOptions = append(cmdLineOptions, "--preview="+previewCmd)

		// Build fzf.Options
		options, err := fzf.ParseOptions(
			true, // whether to load defaults ($FZF_DEFAULT_OPTS_FILE and $FZF_DEFAULT_OPTS)
			cmdLineOptions,
		)
		if err != nil {
			exit(fzf.ExitError, err)
		}

		// Set up input and output channels
		options.Input = inputChan

		// Run fzf
		code, err := fzf.Run(options)
		exit(code, err)
	},
}

func init() {
	rootCmd.AddCommand(pickerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pickerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pickerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
