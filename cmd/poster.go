package cmd

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

// posterCmd represents the poster command
var posterCmd = &cobra.Command{
	Use:   "poster",
	Short: "Given a title, reads poster from db and write it in cwd",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		const filename = "poster.jpeg"
		const yearTolerance = 2
		if _, err := os.ReadFile(filename); err == nil {
			// file is there
			if replace, err := cmd.Flags().GetBool("replace"); err != nil {
				log.Fatalln(" Couldn't read replace flag from config")
			} else if !replace {
				fmt.Printf("Found %v, quitting.\n", filename)
				return
			}
		}
		query := "SELECT poster FROM media WHERE LOWER(media.title)=LOWER(?)"
		title, err := cmd.Flags().GetString("title")
		if err != nil {
			log.Fatalln(" Couldn't read title flag from config")
		}
		if title == "" {
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalln(" Wrong args: title is empty and I can't get the cwd")
			}
			basePath := path.Base(cwd)
			fields := strings.Fields(basePath)
			if len(fields) != 2 {
				log.Fatalf(" Cwd name is badly formatted, must be 'TITLE (YEAR)'")
			}
			title = strings.Join(fields[:len(fields)-1], " ")
		}
		row := config.DBH.DB.QueryRow(query, title)
		var poster []byte
		if err := row.Scan(&poster); err != nil {
			log.Fatalf(" Couldn't get the poster from db for «%v»: %v", title, err)
		}
		img, _, err := image.Decode(bytes.NewReader(poster))
		if err != nil {
			log.Fatalf(" Couldn't decode the poster for «%v»: %v", title, err)
		}
		f, err := os.Create(filename)
		if err != nil {
			log.Fatalf(" Coudln't create the poster file: %v", err)
		}
		defer f.Close()
		if err = jpeg.Encode(f, img, nil); err != nil {
			fmt.Printf(" Failed to encode: %v", err)
		}
		fmt.Printf("✓ Wrote poster to %v\n", filename)
	},
}

func init() {
	rootCmd.AddCommand(posterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// posterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	posterCmd.Flags().BoolP("replace", "r", false, "if replace is true, replace file if it exists")
	posterCmd.Flags().StringP("title", "t", "", "media title, case insensitive, will be read from cwd name if missing")
}
