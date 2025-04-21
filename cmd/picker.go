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
			query := "SELECT title, year, overview, director, id, path FROM media ORDER BY title, year ASC"
			rows, err := localConfig.DBH.DB.Query(query)
			if err != nil {
				log.Fatal("Query error : ", err)
			}
			for rows.Next() {
				var title, overview, director, path string
				var year, id int
				if err := rows.Scan(&title, &year, &overview, &director, &id, &path); err != nil {
					log.Fatal(err)
				}
				if director != "" {
					director = " -- " + director
				}
				infoLine := fmt.Sprintf("%v (%v)%v", title, year, director)
				// TODO: use that folder mod time to influence sorting, see fzf docs
				fileStat, err := os.Stat(path) // WARN: missing records are ignored
				if err != nil {
					s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", title, infoLine, overview, id, path, 0)
					inputChan <- s
				} else {
					s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%s", title, infoLine, overview, id, path, fileStat.ModTime())
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
		cmdLineOptions = append(cmdLineOptions, `--bind=focus:execute-silent(sqlite3 `+localConfig.DBH.Path+` '`+query+`')`)
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
