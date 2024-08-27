package api

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/briandowns/spinner"
)

type Media struct {
	ID               int    `json:"id" validate:"required"`
	Title            string `json:"title" validate:"required_without=Name,omitempty"`
	Name             string `json:"name" validate:"required_without=Title,omitempty"`
	MediaType        string `json:"media_type" validate:"required,oneof=movie tv"`
	ReleaseDate      string `json:"release_date" validate:"required_without=FirstAirDate,omitempty,datetime=2006-01-02"`
	FirstAirDate     string `json:"first_air_date" validate:"required_without=ReleaseDate,omitempty,datetime=2006-01-02"`
	Overview         string `json:"overview"`
	PosterPath       string `json:"poster_path"`
	PosterData       []byte
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	OriginalTitle    string   `json:"original_title"`
	OriginCountry    []string `json:"origin_country"`
	Adult            bool     `json:"adult"`
	Director         string
}

const (
	MediaTypePerson = "person"
	MediaTypeTV     = "tv"
	MediaTypeMovie  = "movie"
)

const CrewJobDirector = "Director"

var MediaTypeIcons = map[string]string{
	MediaTypePerson: "",
	MediaTypeTV:     "",
	MediaTypeMovie:  "󰿏",
}

func (m Media) GetYear() int {
	var dateToParse string
	if m.ReleaseDate == "" {
		dateToParse = m.FirstAirDate
	} else {
		dateToParse = m.ReleaseDate
	}
	date, err := time.Parse(time.DateOnly, dateToParse)
	if err != nil {
		return -1
	}
	return date.Year()
}

func (m Media) String() string {
	return fmt.Sprintf("%v «%v» (%v) @ %v", MediaTypeIcons[m.MediaType], m.GetTitle(), m.GetYear(), m.Url())
}
func (m Media) Dump() string {
	m.PosterData = []byte{}
	out, _ := json.MarshalIndent(m, "", "	")
	return string(out)
}

func (m Media) Url() string {
	return fmt.Sprintf(SiteBaseUrl+"/%v/%v", m.MediaType, m.ID)
}

// GetTitle returns m.Name if m.Title is empty, else return m.Title
func (m Media) GetTitle() string {
	if m.Title == "" {
		return m.Name
	}
	return m.Title
}

// getDirector downloads the first director name in movie credits, if m.MediaType is MediaTypeMovie.
// Will silently use an empty string if m isn't a movie.
func (m *Media) GetDirector(apiReadToken string) {
	if m.MediaType != MediaTypeMovie {
		m.Director = ""
		return
	}
	endpoint := fmt.Sprintf("movie/%v/credits", m.ID)
	data := PollApi(endpoint, "", apiReadToken)
	credits := &MediaCredits{}
	if err := json.Unmarshal(data, credits); err != nil {
		m.Director = ""
		fmt.Printf(" Found no director for %v", m)
		return
	}
	if len(credits.Crew) == 0 {
		m.Director = ""
		fmt.Printf(" Found no director for %v", m)
		return
	}
	firstDirectorIndex := slices.IndexFunc(credits.Crew, func(c CrewMember) bool {
		return c.Job == CrewJobDirector && c.Name != ""
	})
	if firstDirectorIndex == -1 {
		m.Director = ""
		fmt.Printf(" Found no director for %v", m)
		return
	}
	m.Director = credits.Crew[firstDirectorIndex].Name
	fmt.Printf(" Found director %v for %v", m.Director, m)
}

func (m *Media) GetPoster(apiKey string) {
	spinner := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spinner.Suffix = " Downloading poster"
	spinner.Start()
	if m.PosterPath == "" {
		spinner.FinalMSG = " Tried to get the poster of a media without one"
		spinner.Stop()
		return
	}
	spinner.FinalMSG = "✓ Downloaded poster\n"
	data := PollImgApi(m.PosterPath, apiKey)
	spinner.Stop()
	m.PosterData = data
}

type MediaCredits struct {
	Crew []CrewMember
}

type CrewMember struct {
	Job  string
	Name string
}

type MultiSearchResponse struct {
	TotalResults int     `json:"total_results"`
	Results      []Media `json:"results"`
}
