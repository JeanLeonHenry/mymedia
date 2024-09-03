package cmd

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/JeanLeonHenry/mymedia/internal/api"
	"github.com/JeanLeonHenry/mymedia/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

const validationErrorMessage = `Field      %20v
Failed     %20v =%v
Got        %20v (type %v, kind %v)` + "\n"

func parseArgs(cmd *cobra.Command) (string, int, int) {
	title, err := cmd.Flags().GetString("title")
	if err != nil {
		log.Fatalln(" Couldn't read title flag from config")
	}
	year, err := cmd.Flags().GetInt("year")
	if err != nil {
		log.Fatalln(" Couldn't read year flag from config")
	}
	// we assume a year before the invention of cinema or later than 10y in the future is wrong.
	isWrongYear := func(year int) bool { return year <= 1800 || year >= time.Now().Year()+10 }
	if title == "" || isWrongYear(year) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln(" Wrong args: title is empty or year is wrong and I can't get the cwd")
		}
		fmt.Println("Reading info from current dir name")
		basePath := path.Base(cwd)
		fields := strings.Fields(basePath)
		if len(fields) < 2 {
			log.Fatalln(" Cwd name is badly formatted, must be 'TITLE (YEAR)'")
		}
		title = strings.Join(fields[:len(fields)-1], " ")
		lastField := fields[len(fields)-1]
		if currentYear := time.Now().Year(); len(lastField) <= 2 || len(lastField)-2 != len(strconv.Itoa(currentYear)) {
			log.Fatalf(" Year has a wrong amount of digits, its %v and you gave %v\n", currentYear, lastField)
		}
		yearString := lastField[1 : len(lastField)-1]
		year, err = strconv.Atoi(yearString)
		if err != nil {
			log.Fatalln(" Cwd name is badly formatted, must be 'TITLE (YEAR)'")
		} else if isWrongYear(year) {
			log.Fatalf(" Year must be between %v and %v\n", 1800, time.Now().Year()+10)
		}
	}
	tolerance, err := cmd.Flags().GetInt("tolerance")
	if err != nil {
		log.Fatalln(" Couldn't read tolerance flag")
	}
	if tolerance > 5 || tolerance < 0 {
		tolerance = localConfig.DefaultTolerance
		log.Printf(" Tolerance should be between 0 and 5 inclusive. Using %v\n", tolerance)
	}
	return title, year, tolerance
}

// findYearMatch finds the first element of media whose year (given by GetYear()) is minimum.
// If that isn't with tolerance of year, found is false.
// Panics if media is empty.
func findYearMatch(mediaSlice []api.Media, year int, tolerance int) (result api.Media, found bool) {
	distanceToRef := func(x int) int { return utils.Abs(x - year) }
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

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:     "scan",
	Short:   "Scans the current folder for media folders and update database",
	Long:    `WIP`,
	Example: ``,
	Run: func(cmd *cobra.Command, args []string) {
		/*
			PLAN
			1 get (title, year) from config
			2 check db if we have a match: ask if we keep that data (quit) or replace it
			3 poll api, get results, validate them
			4 find a reasonnable match in the results
			5 check db before writing the match if user accepts
		*/
		// 1
		title, year, tolerance := parseArgs(cmd)
		// 2
		if localConfig.DBH.CheckDB(title, year, tolerance, debug) {
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
		localConfig.DBH.CheckDB(media.GetTitle(), media.GetYear(), tolerance, debug)
		utils.AcceptOrQuit("Write to DB ?")
		media.GetDirector(localConfig.ApiReadToken)
		media.GetPoster(localConfig.ApiKey)
		if cwdPath, err := os.Getwd(); err != nil {
			log.Fatalln(" Couldn't get current dir path")
		} else {
			_, err := localConfig.DBH.WriteToDB(media, cwdPath)
			if err != nil {
				log.Fatalln(" DB write error: ", err)
			}
			fmt.Println("✓ Wrote to DB: ", media)
			if debug {
				fmt.Println("Tried writing/Wrote: ", media.Dump())
			}
		}
		if debug {
			fmt.Println("-- DUMP --")
			fmt.Println("Dumping config")
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
