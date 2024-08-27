package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

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
			query := "SELECT title, year, overview, director, path FROM media ORDER BY title, year ASC"
			rows, err := localConfig.DBH.DB.Query(query)
			if err != nil {
				log.Fatal("Query error : ", err)
			}
			for rows.Next() {
				var title, overview, director, path string
				var year int
				if err := rows.Scan(&title, &year, &overview, &director, &path); err != nil {
					log.Fatal(err)
				}
				if director != "" {
					director = " -- " + director
				}
				s := fmt.Sprintf("%v\t%v (%v)%v\t%v\t%v", title, title, year, director, overview, path)
				inputChan <- s
			}
			close(inputChan)
		}()

		outputChan := make(chan string)
		go func() {
			for s := range outputChan {
				path := strings.FieldsFunc(s, func(r rune) bool { return r == '\t' })[3]
				fmt.Println(path)
			}
		}()

		exit := func(code int, err error) {
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			os.Exit(code)
		}

		cmdLineOptions := []string{"--delimiter=\\t", "--with-nth=1"}
		// TODO: use poster image from db. idea: when fzf selected item change, write poster blob to a tmp file, use that in preview
		previewCmd := "echo {2};echo;echo {3}|fold -w ${FZF_PREVIEW_COLUMNS} -s;COLS=$((LINES*2/3));kitten icat --clear --transfer-mode=memory --stdin=no --unicode-placeholder --place=${COLS}x${FZF_PREVIEW_LINES}@0x0 {-1}/poster.*"
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
		options.Output = outputChan

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
