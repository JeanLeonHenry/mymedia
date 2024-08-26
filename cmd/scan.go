package cmd

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

var abs = func(x int) int {
	if x <= 0 {
		return -x
	}
	return x
}

// TODO: add validation tags
type Media struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Name             string   `json:"name"`
	MediaType        string   `json:"media_type"`
	ReleaseDate      string   `json:"release_date" validate:"required_without=FirstAirDate,datetime=2006-01-02"`
	FirstAirDate     string   `json:"first_air_date" validate:"required_without=ReleaseDate,datetime=2006-01-02"`
	Overview         string   `json:"overview"`
	PosterPath       any      `json:"poster_path"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	OriginalTitle    string   `json:"original_title"`
	OriginCountry    []string `json:"origin_country"`
	Adult            bool     `json:"adult"`
}

const (
	MediaTypePerson = "person"
	MediaTypeTV     = "tv"
	MediaTypeMovie  = "movie"
)

var MediaTypeIcons = map[string]string{
	MediaTypePerson: "",
	MediaTypeTV:     "",
	MediaTypeMovie:  "󰿏",
}

func (m Media) GetYear() int {
	// NOTE: this is assumed to take place after validation, so at least one is non empty and parses out to a date
	var dateToParse string
	if m.ReleaseDate == "" {
		dateToParse = m.FirstAirDate
	} else {
		dateToParse = m.ReleaseDate
	}
	date, _ := time.Parse(time.DateOnly, dateToParse)
	return date.Year()
}

func (m Media) Url() string {
	return fmt.Sprintf("https://themoviedb.org/%v/%v", m.MediaType, m.ID)
}

func (m Media) GetTitle() string {
	// NOTE: this is assumed to take place after validation, so at least of the two is non empty
	if m.Title == "" {
		return m.Name
	}
	return m.Title
}

func (m Media) writeToDB() {
	// TODO: implement write data to db
}

func (m Media) downloadExtraInfo() Media {
	// TODO: implement director and poster download
	return m
}

func (m Media) String() (result string) {
	result = fmt.Sprintf("%v «%v» (%v) @ %v", MediaTypeIcons[m.MediaType], m.GetTitle(), m.GetYear(), m.Url())
	return
}

type MultiSearchResponse struct {
	TotalResults int     `json:"total_results"`
	Results      []Media `json:"results"`
}

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
			log.Fatalf(" Cwd name is badly formatted, must be 'TITLE (YEAR)'")
		}
		title = strings.Join(fields[:len(fields)-1], " ")
		lastField := fields[len(fields)-1]
		if currentYear := time.Now().Year(); len(lastField) <= 2 || len(lastField)-2 != len(strconv.Itoa(currentYear)) {
			log.Fatalf(" Year has a wrong amount of digits, its %v and you gave %v", currentYear, lastField)
		}
		yearString := lastField[1 : len(lastField)-1]
		year, err = strconv.Atoi(yearString)
		if err != nil {
			log.Fatalf(" Cwd name is badly formatted, must be 'TITLE (YEAR)'")
		} else if isWrongYear(year) {
			log.Fatalf(" Year must be between %v and %v", 1800, time.Now().Year()+10)
		}
	}
	tolerance, err := cmd.Flags().GetInt("tolerance")
	if err != nil {
		log.Fatalf(" Couldn't read tolerance flag")
	}
	if tolerance > 5 || tolerance < 0 {
		tolerance = config.DefaultTolerance
		log.Printf(" Tolerance should be between 0 and 5 inclusive. Using %v", tolerance)
	}
	return title, year, tolerance
}

func acceptOrQuit(prompt string) {
	fmt.Print(prompt + " [y/N] ")
	var userInput string
	if _, err := fmt.Scanln(&userInput); err != nil || userInput != "y" {
		fmt.Print("Quitting.")
		os.Exit(1)
	}
}

// checkDB looks up the db for a media record with case-insensitive matching titles and a year within tolerance of year
func checkDB(title string, year int, tolerance int) bool {
	dBQuery := "SELECT title, year, id, media_type FROM media WHERE lower(media.title)=lower(?) AND ABS(media.year-?)<=?"
	rows := config.DBH.DB.QueryRow(dBQuery, title, year, tolerance)
	var titleDB, media_type string
	var yearDB, id int
	if err := rows.Scan(&titleDB, &yearDB, &id, &media_type); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Fatal(" Query error: ", err)
		}
		// found no match, check is complete
		return false
	}
	fmt.Printf("✓ Found %v (%v) in DB.\n Its TMDB page is https://www.themoviedb.org/%v/%v\n", titleDB, yearDB, media_type, id)
	return true
}

func apiMultiSearch(apiQuery string) MultiSearchResponse {
	endpoint := "search/multi"
	if apiQuery == "" {
		fmt.Printf("󰍉 Searching TMDB.org on %v\n", endpoint)
	} else {
		fmt.Printf("󰍉 Searching TMDB.org on %v for %v\n", endpoint, apiQuery)
	}
	fullUrl, err := url.JoinPath(config.ApiUrl, endpoint)
	if err != nil {
		log.Fatalf(" Api url couldn't be formed: %+v", config)
	}
	parsedUrl, err := url.Parse(fullUrl)
	if err != nil {
		log.Fatalln(" Api url couldn't be parsed: check config file")
	}
	if apiQuery != "" {
		v := url.Values{}
		v.Set("query", apiQuery)
		parsedUrl.RawQuery = v.Encode()
	}
	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.ApiReadToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(" Error contacting the API: ", err)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf(" Error reading the API's response")
	}
	if resp.StatusCode >= 400 {
		log.Fatalln(" Error contacting the API: ", data)
	}

	object := new(MultiSearchResponse)
	if err := json.Unmarshal(data, object); err != nil {
		log.Fatalln(" Error unpacking the API's response: ", err)
	}
	if *config.Debug {
		// log.Printf("Got results: %+v\nfrom request %+v", object.Results, req)
	}
	return *object
}

// findYearMatch finds the first element of media whose year (given by GetYear()) is minimum.
// If that isn't with tolerance of year, found is false.
// Panics if media is empty.
func findYearMatch(media []Media, year int, tolerance int) (result Media, found bool) {
	distanceToRef := func(x int) int { return abs(x - year) }
	result = slices.MinFunc(media, func(a, b Media) int {
		yearA, yearB := a.GetYear(), b.GetYear()
		return cmp.Compare(distanceToRef(yearA), distanceToRef(yearB))
	})
	if distanceToRef(result.GetYear()) > tolerance {
		return result, false
	}
	return result, true
}

func validateResults(validate *validator.Validate, results []Media) (validResults []Media) {
	for _, r := range results {
		err := validate.Struct(r)
		if err != nil {
			if !*config.Debug {
				continue
			}
			for _, err := range err.(validator.ValidationErrors) {
				// TODO: better validation error msg
				fmt.Println(" Error validating: " + fmt.Sprintf("%+v", r))
				fmt.Println("Field ", err.StructNamespace(), err.StructField())
				fmt.Printf("Got %v\n", err.Value())
				fmt.Println("Validation tag ", err.Tag(), err.ActualTag(), "with param", err.Param())
				fmt.Println("Datatypes ", err.Kind(), err.Type())
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
		// TODO: remove this warning when finished
		if !*config.Debug {
			fmt.Println(" this is not done. Use --debug/-d flag to continue development, or the python script for actual use.")
		}
		/*
			PLAN
			1 get (title, year) from config
			2 check db if we have a match: ask if we keep that data (quit) or replace it
			3 poll api, get results, validate them
			4 find a reasonnable match in the results
			5 check db before writing the match
		*/
		title, year, tolerance := parseArgs(cmd)
		if checkDB(title, year, tolerance) {
			acceptOrQuit("Proceed to online lookup?")
		}
		response := apiMultiSearch(title)
		validate := validator.New(validator.WithRequiredStructEnabled())
		validResults := validateResults(validate, response.Results)
		if len(validResults) == 0 {
			fmt.Printf("∅ Found no match for «%v» (%v).\n", title, year)
			return
		}
		media, ok := findYearMatch(validResults, year, tolerance)
		if !ok {
			fmt.Printf("∅ Found no match for «%v» (%v).\nClosest match was : %+v\n", title, year, media)
			return
		}
		fmt.Printf("✓ Found TMDB.org match for «%v» (%v): %v\n", title, year, media.Url())
		if checkDB(media.GetTitle(), media.GetYear(), tolerance) {
			acceptOrQuit("Write to DB anyway?")
		}
		media.downloadExtraInfo().writeToDB()
	},
}

func init() {
	log.Default().SetFlags(log.LstdFlags | log.Lshortfile)
	rootCmd.AddCommand(scanCmd)
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	scanCmd.Flags().BoolP("replace", "r", false, "if replace is true, replace db data if it exists")
	scanCmd.Flags().StringP("title", "t", "", "media title, case insensitive, will be read from cwd name if missing")
	scanCmd.Flags().IntP("year", "y", 0, "media release year")
	scanCmd.Flags().Int("tolerance", 2, "on lookup, result will be accepted if title match and year is within tolerance of result")
	scanCmd.Flags().BoolVarP(config.Debug, "debug", "d", false, "if true, use wip implementation in go, else use working one in python")

}
