package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/JeanLeonHenry/mymedia/db"
	"github.com/JeanLeonHenry/mymedia/internal/utils"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove db items with wrong paths.",
	Run: func(cmd *cobra.Command, args []string) {
		results, err := queries.ListPaths(ctx)
		if err != nil {
			log.Fatalln(err)
		}
		for _, result := range results {
			if _, err := os.Stat(result.Path); err != nil {
				fmt.Println(utils.Bad("Path %v is not valid : %v.", result.Path, err))
				replacementPath, err := utils.AskUserForAPath("Provide a replacement path : ")
				if err == nil {
					err := queries.UpdatePath(ctx, db.UpdatePathParams{
						Path: replacementPath,
						ID:   result.ID,
					})
					if err != nil {
						log.Printf("Couldn't update the path at media id %v\nGot : %v", result.ID, err)
					}
					continue
				}
				fmt.Println(utils.Bad("%v", err))
				forceFlag, _ := cmd.Flags().GetBool("force")
				doIt := true
				if !forceFlag {
					doIt = utils.AskUser(fmt.Sprintf("Delete media at path %v ?", result.Path))
				}
				if doIt {
					err := queries.DeleteMedia(ctx, result.Path)
					if err != nil {
						log.Println(err)
					} else {
						fmt.Println(utils.Good("Deleted media at path %v", result.Path))
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cleanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	cleanCmd.Flags().BoolP("force", "f", false, "Don't ask for confirmation before deleting from db.")
}
