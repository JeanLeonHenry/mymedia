package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

const SearchMultiEndPoint = "search/multi"
const MovieEndpointPattern = `movie/\d+/credits`
const SiteBaseUrl = "https://themoviedb.org"
const ApiBaseUrl = "https://api.themoviedb.org/3"
const ImgApiBaseUrl = "http://image.tmdb.org/t/p/w500"

func formUrl(baseUrl, endpoint string) string {
	fullUrl, err := url.JoinPath(baseUrl, endpoint)
	if err != nil {
		log.Fatalf(" Api url couldn't be formed:\nBase url: %v\nEndpoint: %v\n", baseUrl, endpoint)
	}
	return fullUrl
}
func HTTPResponseBodyData(resp *http.Response, err error) []byte {
	if err != nil {
		log.Fatalf(" Error contacting the API: %v", err)
	}
	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf(" Error reading the API's response")
	}
	if resp.StatusCode >= 400 {
		log.Fatalln(" Api response error: ", data)
	}
	return data
}
func PollApi(endpoint, apiQuery, apiReadToken string) []byte {
	if isMovieCreditsEndpoint, _ := regexp.MatchString(MovieEndpointPattern, endpoint); endpoint != SearchMultiEndPoint && !isMovieCreditsEndpoint {
		log.Fatalf(" Api request error: endpoint was %v", endpoint)
	}
	if apiQuery == "" {
		fmt.Printf("󰍉 Searching TMDB.org on %v\n", endpoint)
	} else {
		fmt.Printf("󰍉 Searching TMDB.org on %v for %v\n", endpoint, apiQuery)
	}
	fullUrl := formUrl(ApiBaseUrl, endpoint)
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
	req.Header.Set("Authorization", "Bearer "+apiReadToken)
	resp, err := http.DefaultClient.Do(req)
	return HTTPResponseBodyData(resp, err)
}
func PollImgApi(posterPath, apiKey string) []byte {
	fmt.Printf("󰍉 Downloading poster @ %v", posterPath)
	fullUrl := formUrl(ImgApiBaseUrl, posterPath)
	parsedUrl, err := url.Parse(fullUrl)
	if err != nil {
		log.Fatalln(" Api url couldn't be parsed: check config file")
	}
	v := url.Values{}
	v.Set("api_key", apiKey)
	parsedUrl.RawQuery = v.Encode()
	resp, err := http.Get(parsedUrl.String())
	return HTTPResponseBodyData(resp, err)
}
func ApiMultiSearch(apiQuery, apiReadToken string) MultiSearchResponse {
	data := PollApi(SearchMultiEndPoint, apiQuery, apiReadToken)
	object := new(MultiSearchResponse)
	if err := json.Unmarshal(data, object); err != nil {
		log.Fatalln(" Error unpacking the API's response: ", err)
	}
	return *object
}
