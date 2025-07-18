package cmd

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/JeanLeonHenry/mymedia/db"
	"github.com/JeanLeonHenry/mymedia/internal/api"
	"github.com/JeanLeonHenry/mymedia/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

const validationErrorMessage = `Field      %20v
Failed     %20v =%v
Got        %20v (type %v, kind %v)` + "\n"

func currentDirFormatError(extra string) {
	log.Fatalln(" Cwd name is badly formatted, must be 'TITLE (YEAR) [tmdbid-ID]'. The tmdbid field is optional.", extra)
}

// parseArgs returns the title, year, year tolerance and tmdbid.
// title and year are first tried from flags.
// If no title can be found from flags or year is wrong, try from cwd name.
// If no tmdbid is found in cwd name, returns 0.
func parseArgs(cmd *cobra.Command) (string, int, int, int) {
	var tmdbId int
	title, err := cmd.Flags().GetString("title")
	if err != nil {
		log.Fatalln(" Couldn't read title flag from config")
	}
	year, err := cmd.Flags().GetInt("year")
	if err != nil {
		log.Fatalln(" Couldn't read year flag from config")
	}
	tolerance, err := cmd.Flags().GetInt("tolerance")
	if err != nil {
		log.Fatalln(" Couldn't read tolerance flag")
	}
	if tolerance > 5 || tolerance < 0 {
		tolerance = localConfig.DefaultTolerance
		log.Printf(" Tolerance should be between 0 and 5 inclusive. Using %v\n", tolerance)
	}
	// we assume a year before the invention of cinema or later than 10y in the future is wrong.
	isWrongYear := func(year int) bool { return year <= 1800 || year >= time.Now().Year()+10 }
	if title == "" || isWrongYear(year) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln(" Wrong args: title is empty or year is wrong, and I can't get the cwd")
		}
		fmt.Println("Reading info from current folder name")
		basePath := path.Base(cwd)
		var re = regexp.MustCompile(`(.*) \((\d{4})\)( \[tmdbid-(\d+)\])?`)
		fields := re.FindStringSubmatch(basePath)
		if fields == nil || len(fields) <= 1 {
			currentDirFormatError("")
		}
		title = fields[1]
		yearString := fields[2]
		if len(fields) == 5 && fields[4] != "" {
			tmdbId, err = strconv.Atoi(fields[4])
			if err != nil {
				currentDirFormatError(fmt.Sprintf("Couldn't parse tmdbid '%v' to an int. Parsing provided fields %v", fields[4], fields))
			}
		}
		year, err = strconv.Atoi(yearString)
		if err != nil {
			currentDirFormatError(fmt.Sprintf("Couldn't parse year %v to an int", yearString))
		} else if isWrongYear(year) {
			log.Fatalf(" Year must be between %v and %v\n", 1800, time.Now().Year()+10)
		}
	}
	return title, year, tolerance, tmdbId
}

// findYearMatch finds the first element of media whose year (given by GetYear()) is minimum.
// If that isn't with tolerance of year, found is false.
// Panics if media is empty.
func findYearMatch(mediaSlice []api.Media, year int, tolerance int) (result api.Media, found bool) {
	distanceToRef := func(x int) int { return utils.Abs(x - year) }
	// INFO: panics if mediaSlice is empty
	result = slices.MinFunc(mediaSlice, func(a, b api.Media) int {
		yearA, yearB := a.GetYear(), b.GetYear()
		return cmp.Compare(distanceToRef(yearA), distanceToRef(yearB))
	})
	if distanceToRef(result.GetYear()) > tolerance {
		return result, false
	}
	return result, true
}

func validateResults(validate *validator.Validate, results []api.Media) (validResults []api.Media) {
	for _, r := range results {
		err := validate.Struct(r)
		if err != nil {
			if !debug {
				continue
			}
			for _, err := range err.(validator.ValidationErrors) {
				fmt.Println(" Error validating result: " + r.Dump())
				fmt.Printf(validationErrorMessage,
					err.StructNamespace(),
					err.ActualTag(), err.Param(),
					err.Value(),
					err.Type(),
					err.Kind(),
				)
				os.Exit(1)
			}
		}
		validResults = append(validResults, r)
	}
	return
}

// checkDB looks for a media in db with same title and year within tolerance of given year.
// title are compared in lowercase
func checkDB(q *db.Queries, c context.Context, title string, year int, tolerance int, debug bool) bool {
	results, err := q.LookUpMedia(c, db.LookUpMediaParams{
		Title:     title,
		Year:      int64(year),
		Tolerance: int64(tolerance),
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Fatal(" Query error: ", err)
		}
		// found no match, check is complete
		return false
	}
	for result := range results {
		fmt.Printf("✓ Found %v in DB.\n", result)
	}
	return true
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scans the current folder for media folders and update database",
	Long:  ``,
	Example: `$ mymedia scan
will get media info from current directory and proceed to look up.

If the result is wrong, use the -t and -y flags to make lookup more accurate, especially for foreign movies.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		/*
			PLAN
			1 get (title, year) from config
			2 check db if we have a match: ask if we keep that data (quit) or replace it
			3 poll api, get results, validate them
			4 find a reasonnable match in the results
			5 check db before writing the match if user accepts
		*/
		// FIX: if the title contains a tmdbid, it should be used in the API poll

		// 1
		title, year, tolerance, _ := parseArgs(cmd)
		// 2
		if checkDB(queries, ctx, title, year, tolerance, debug) {
			utils.AcceptOrQuit("Proceed to online lookup?")
		}
		// 3
		response := api.ApiMultiSearch(title, localConfig.ApiReadToken)
		validate := validator.New(validator.WithRequiredStructEnabled())
		validResults := validateResults(validate, response.Results)
		if len(validResults) == 0 {
			fmt.Printf("∅ Found no match for «%v» (%v).\n", title, year)
			return
		}
		// 4
		media, ok := findYearMatch(validResults, year, tolerance)
		if !ok {
			out := media.String()
			if debug {
				out = media.Dump()
			}
			fmt.Printf("∅ Found no match for «%v» (%v).\nClosest match was : %+v\n", title, year, out)
			return
		}
		out := media.String()
		if debug {
			out = media.Dump()
		}
		fmt.Printf("✓ Found TMDB.org match for «%v» (%v): %v\n", title, year, out)
		// 5
		checkDB(queries, ctx, media.GetTitle(), media.GetYear(), tolerance, debug)
		utils.AcceptOrQuit("Write to DB ?")
		media.GetDirector(localConfig.ApiReadToken)
		media.GetPoster(localConfig.ApiKey)
		if cwdPath, err := os.Getwd(); err == nil {
			err := queries.InsertOrReplaceMedia(ctx, db.InsertOrReplaceMediaParams{
				ID:        int64(media.ID),
				MediaType: media.MediaType,
				Title:     media.GetTitle(),
				Year:      int64(media.GetYear()),
				Overview: sql.NullString{
					String: media.Overview,
					Valid:  media.Overview != "",
				},
				Director: sql.NullString{
					String: media.Director,
					Valid:  media.Director != "",
				},
				Poster: media.PosterData,
				Path:   cwdPath,
			})
			if err != nil {
				log.Fatalln(" DB write error: ", err)
			}
			fmt.Println("✓ Wrote to DB: ", media)
			if debug {
				fmt.Println("Tried writing/Wrote: ", media.Dump())
			}
		} else {
			log.Fatalln(" Couldn't get current dir path")
		}
		if debug {
			fmt.Println("-- CONFIG DUMP --")
			fmt.Println(localConfig)
		}
	},
}

func init() {
	if debug {
		log.Default().SetFlags(log.LstdFlags | log.Lshortfile)
	}
	rootCmd.AddCommand(scanCmd)
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	scanCmd.Flags().StringP("title", "t", "", "media title, case insensitive, will be read from cwd name if missing")
	scanCmd.Flags().IntP("year", "y", 0, "media release year")
	scanCmd.Flags().Int("tolerance", 2, "on lookup, result will be accepted if title match and year is within tolerance of result")

}
