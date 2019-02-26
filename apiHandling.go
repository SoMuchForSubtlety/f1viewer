package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const urlStart = "https://f1tv.formula1.com"

const tagsURL = "https://f1tv.formula1.com/api/tags/"
const vodTypesURL = "http://f1tv.formula1.com/api/vod-type-tag/"
const seriesListURL = "https://f1tv.formula1.com/api/series/"
const seriesF1URL = "https://f1tv.formula1.com/api/series/seri_436bb431c3a24d7d8e200a74e1d11de4/"
const teamsURL = "https://f1tv.formula1.com/api/episodes/"

type episodeStruct struct {
	Subtitle               string        `json:"subtitle"`
	UID                    string        `json:"uid"`
	ScheduleUrls           []string      `json:"schedule_urls"`
	SessionoccurrenceUrls  []string      `json:"sessionoccurrence_urls"`
	Stats                  interface{}   `json:"stats"`
	Title                  string        `json:"title"`
	Self                   string        `json:"self"`
	DriverUrls             []string      `json:"driver_urls"`
	CircuitUrls            []interface{} `json:"circuit_urls"`
	VodTypeTagUrls         []string      `json:"vod_type_tag_urls"`
	DataSourceFields       []interface{} `json:"data_source_fields"`
	ParentURL              interface{}   `json:"parent_url"`
	DataSourceID           string        `json:"data_source_id"`
	Tags                   []interface{} `json:"tags"`
	ImageUrls              []string      `json:"image_urls"`
	SeriesUrls             []interface{} `json:"series_urls"`
	TeamUrls               []string      `json:"team_urls"`
	HierarchyURL           string        `json:"hierarchy_url"`
	SponsorUrls            []interface{} `json:"sponsor_urls"`
	PlanUrls               []interface{} `json:"plan_urls"`
	EpisodeNumber          interface{}   `json:"episode_number"`
	Slug                   string        `json:"slug"`
	LastDataIngest         time.Time     `json:"last_data_ingest"`
	Talent                 []interface{} `json:"talent"`
	Language               string        `json:"language"`
	Created                time.Time     `json:"created"`
	Items                  []string      `json:"items"`
	RatingUrls             []interface{} `json:"rating_urls"`
	Modified               time.Time     `json:"modified"`
	RecommendedContentUrls []interface{} `json:"recommended_content_urls"`
	Synopsis               string        `json:"synopsis"`
	Editability            string        `json:"editability"`
}

type assetStruct struct {
	MaxDevices             interface{}   `json:"max_devices"`
	UID                    string        `json:"uid"`
	ScheduleUrls           []string      `json:"schedule_urls"`
	Self                   string        `json:"self"`
	SessionoccurrenceUrls  []interface{} `json:"sessionoccurrence_urls"`
	Duration               string        `json:"duration"`
	Stats                  interface{}   `json:"stats"`
	Title                  string        `json:"title"`
	Guidance               bool          `json:"guidance"`
	AssetTypeURL           interface{}   `json:"asset_type_url"`
	DriverUrls             []interface{} `json:"driver_urls"`
	CircuitUrls            []interface{} `json:"circuit_urls"`
	DurationInSeconds      int           `json:"duration_in_seconds"`
	Subtitles              bool          `json:"subtitles"`
	DataSourceFields       []string      `json:"data_source_fields"`
	ParentURL              string        `json:"parent_url"`
	DataSourceID           string        `json:"data_source_id"`
	VodTypeTagUrls         []interface{} `json:"vod_type_tag_urls"`
	StatsLastUpdated       interface{}   `json:"stats_last_updated"`
	Tags                   []interface{} `json:"tags"`
	GuidanceText           string        `json:"guidance_text"`
	AccountUrls            []string      `json:"account_urls"`
	SeriesUrls             []interface{} `json:"series_urls"`
	TeamUrls               []interface{} `json:"team_urls"`
	HierarchyURL           string        `json:"hierarchy_url"`
	SponsorUrls            []interface{} `json:"sponsor_urls"`
	ImageUrls              []interface{} `json:"image_urls"`
	PlanUrls               []interface{} `json:"plan_urls"`
	Slug                   string        `json:"slug"`
	LastDataIngest         time.Time     `json:"last_data_ingest"`
	Sound                  bool          `json:"sound"`
	Talent                 []interface{} `json:"talent"`
	Language               string        `json:"language"`
	Created                time.Time     `json:"created"`
	URL                    string        `json:"url"`
	ReleaseDate            interface{}   `json:"release_date"`
	RatingUrls             []interface{} `json:"rating_urls"`
	Modified               time.Time     `json:"modified"`
	RecommendedContentUrls []interface{} `json:"recommended_content_urls"`
	Ovps                   []struct {
		AccountURL string `json:"account_url"`
		StreamURL  string `json:"stream_url"`
	} `json:"ovps"`
	Licensor    string `json:"licensor"`
	Editability string `json:"editability"`
}

type seriesStruct struct {
	Name                  string        `json:"name"`
	Language              string        `json:"language"`
	Created               time.Time     `json:"created"`
	Self                  string        `json:"self"`
	Modified              time.Time     `json:"modified"`
	ImageUrls             []interface{} `json:"image_urls"`
	ContentUrls           []interface{} `json:"content_urls"`
	LastDataIngest        time.Time     `json:"last_data_ingest"`
	DataSourceFields      []string      `json:"data_source_fields"`
	SessionoccurrenceUrls []string      `json:"sessionoccurrence_urls"`
	Editability           string        `json:"editability"`
	DataSourceID          string        `json:"data_source_id"`
	UID                   string        `json:"uid"`
}

type vodTypesStruct struct {
	Objects []struct {
		Name             string        `json:"name"`
		Language         string        `json:"language"`
		Created          time.Time     `json:"created"`
		Self             string        `json:"self"`
		Modified         time.Time     `json:"modified"`
		ImageUrls        []interface{} `json:"image_urls"`
		ContentUrls      []string      `json:"content_urls"`
		LastDataIngest   time.Time     `json:"last_data_ingest"`
		DataSourceFields []string      `json:"data_source_fields"`
		Editability      string        `json:"editability"`
		DataSourceID     string        `json:"data_source_id"`
		UID              string        `json:"uid"`
	} `json:"objects"`
}

//downloads json from URL and returns the json as string and whether it's valid as bool
func getJSON(url string) (bool, string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	response := buf.String()
	return isJSON(response), response
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func getEpisode(episodeID string) episodeStruct {
	var ep episodeStruct

	_, jsonString := getJSON(urlStart + episodeID)
	json.Unmarshal([]byte(jsonString), &ep)

	return ep
}

func getVodTypes() vodTypesStruct {
	var types vodTypesStruct

	_, jsonString := getJSON(vodTypesURL)
	json.Unmarshal([]byte(jsonString), &types)

	return types
}
