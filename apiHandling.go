package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const urlStart = "https://f1tv.formula1.com"
const sessionURLstart = "https://f1tv.formula1.com/api/session-occurrence/?fields=uid,nbc_status,status,editorial_start_time,live_sources_path,data_source_id,available_for_user,global_channel_urls,global_channel_urls__uid,global_channel_urls__slug,global_channel_urls__self,channel_urls,channel_urls__ovps,channel_urls__slug,channel_urls__name,channel_urls__uid,channel_urls__self,channel_urls__driver_urls,channel_urls__driver_urls__driver_tla,channel_urls__driver_urls__driver_racingnumber,channel_urls__driver_urls__first_name,channel_urls__driver_urls__last_name,channel_urls__driver_urls__image_urls,channel_urls__driver_urls__image_urls__image_type,channel_urls__driver_urls__image_urls__url,channel_urls__driver_urls__team_url,channel_urls__driver_urls__team_url__name,channel_urls__driver_urls__team_url__colour,eventoccurrence_url,eventoccurrence_url__slug,eventoccurrence_url__circuit_url,eventoccurrence_url__circuit_url__short_name,session_type_url,session_type_url__name&fields_to_expand=global_channel_urls,channel_urls,channel_urls__driver_urls,channel_urls__driver_urls__image_urls,channel_urls__driver_urls__team_url,eventoccurrence_url,eventoccurrence_url__circuit_url,session_type_url&slug="

const tagsURL = "https://f1tv.formula1.com/api/tags/"
const vodTypesURL = "http://f1tv.formula1.com/api/vod-type-tag/"
const seriesListURL = "https://f1tv.formula1.com/api/series/"
const seriesF1URL = "https://f1tv.formula1.com/api/series/seri_436bb431c3a24d7d8e200a74e1d11de4/"
const teamsURL = "https://f1tv.formula1.com/api/episodes/"

type episodeStruct struct {
	Subtitle               string    `json:"subtitle"`
	UID                    string    `json:"uid"`
	ScheduleUrls           []string  `json:"schedule_urls"`
	SessionoccurrenceUrls  []string  `json:"sessionoccurrence_urls"`
	Stats                  string    `json:"stats"`
	Title                  string    `json:"title"`
	Self                   string    `json:"self"`
	DriverUrls             []string  `json:"driver_urls"`
	CircuitUrls            []string  `json:"circuit_urls"`
	VodTypeTagUrls         []string  `json:"vod_type_tag_urls"`
	DataSourceFields       []string  `json:"data_source_fields"`
	ParentURL              string    `json:"parent_url"`
	DataSourceID           string    `json:"data_source_id"`
	Tags                   []string  `json:"tags"`
	ImageUrls              []string  `json:"image_urls"`
	SeriesUrls             []string  `json:"series_urls"`
	TeamUrls               []string  `json:"team_urls"`
	HierarchyURL           string    `json:"hierarchy_url"`
	SponsorUrls            []string  `json:"sponsor_urls"`
	PlanUrls               []string  `json:"plan_urls"`
	EpisodeNumber          string    `json:"episode_number"`
	Slug                   string    `json:"slug"`
	LastDataIngest         time.Time `json:"last_data_ingest"`
	Talent                 []string  `json:"talent"`
	Language               string    `json:"language"`
	Created                time.Time `json:"created"`
	Items                  []string  `json:"items"`
	RatingUrls             []string  `json:"rating_urls"`
	Modified               time.Time `json:"modified"`
	RecommendedContentUrls []string  `json:"recommended_content_urls"`
	Synopsis               string    `json:"synopsis"`
	Editability            string    `json:"editability"`
}

type assetStruct struct {
	MaxDevices             interface{}   `json:"max_devices"`
	UID                    string        `json:"uid"`
	ScheduleUrls           []string      `json:"schedule_urls"`
	Self                   string        `json:"self"`
	SessionoccurrenceUrls  []string      `json:"sessionoccurrence_urls"`
	Duration               string        `json:"duration"`
	Stats                  interface{}   `json:"stats"`
	Title                  string        `json:"title"`
	Guidance               bool          `json:"guidance"`
	AssetTypeURL           string        `json:"asset_type_url"`
	DriverUrls             []string      `json:"driver_urls"`
	CircuitUrls            []string      `json:"circuit_urls"`
	DurationInSeconds      int           `json:"duration_in_seconds"`
	Subtitles              bool          `json:"subtitles"`
	DataSourceFields       []string      `json:"data_source_fields"`
	ParentURL              string        `json:"parent_url"`
	DataSourceID           string        `json:"data_source_id"`
	VodTypeTagUrls         []string      `json:"vod_type_tag_urls"`
	StatsLastUpdated       interface{}   `json:"stats_last_updated"`
	Tags                   []interface{} `json:"tags"`
	GuidanceText           string        `json:"guidance_text"`
	AccountUrls            []string      `json:"account_urls"`
	SeriesUrls             []string      `json:"series_urls"`
	TeamUrls               []string      `json:"team_urls"`
	HierarchyURL           string        `json:"hierarchy_url"`
	SponsorUrls            []string      `json:"sponsor_urls"`
	ImageUrls              []string      `json:"image_urls"`
	PlanUrls               []string      `json:"plan_urls"`
	Slug                   string        `json:"slug"`
	LastDataIngest         time.Time     `json:"last_data_ingest"`
	Sound                  bool          `json:"sound"`
	Talent                 []interface{} `json:"talent"`
	Language               string        `json:"language"`
	Created                time.Time     `json:"created"`
	URL                    string        `json:"url"`
	ReleaseDate            interface{}   `json:"release_date"`
	RatingUrls             []string      `json:"rating_urls"`
	Modified               time.Time     `json:"modified"`
	RecommendedContentUrls []string      `json:"recommended_content_urls"`
	Ovps                   []struct {
		AccountURL string `json:"account_url"`
		StreamURL  string `json:"stream_url"`
	} `json:"ovps"`
	Licensor    string `json:"licensor"`
	Editability string `json:"editability"`
}

type seriesStruct struct {
	Name                  string    `json:"name"`
	Language              string    `json:"language"`
	Created               time.Time `json:"created"`
	Self                  string    `json:"self"`
	Modified              time.Time `json:"modified"`
	ImageUrls             []string  `json:"image_urls"`
	ContentUrls           []string  `json:"content_urls"`
	LastDataIngest        time.Time `json:"last_data_ingest"`
	DataSourceFields      []string  `json:"data_source_fields"`
	SessionoccurrenceUrls []string  `json:"sessionoccurrence_urls"`
	Editability           string    `json:"editability"`
	DataSourceID          string    `json:"data_source_id"`
	UID                   string    `json:"uid"`
}

type vodTypesStruct struct {
	Objects []struct {
		Name             string    `json:"name"`
		Language         string    `json:"language"`
		Created          time.Time `json:"created"`
		Self             string    `json:"self"`
		Modified         time.Time `json:"modified"`
		ImageUrls        []string  `json:"image_urls"`
		ContentUrls      []string  `json:"content_urls"`
		LastDataIngest   time.Time `json:"last_data_ingest"`
		DataSourceFields []string  `json:"data_source_fields"`
		Editability      string    `json:"editability"`
		DataSourceID     string    `json:"data_source_id"`
		UID              string    `json:"uid"`
	} `json:"objects"`
}

type driverStruct struct {
	LastName                     string    `json:"last_name"`
	UID                          string    `json:"uid"`
	EventoccurrenceAsWinner1Urls []string  `json:"eventoccurrence_as_winner_1_urls"`
	NationURL                    string    `json:"nation_url"`
	ChannelUrls                  []string  `json:"channel_urls"`
	LastSeason                   int       `json:"last_season"`
	FirstName                    string    `json:"first_name"`
	DriverReference              string    `json:"driver_reference"`
	Self                         string    `json:"self"`
	FirstSeason                  int       `json:"first_season"`
	DriverTla                    string    `json:"driver_tla"`
	DataSourceFields             []string  `json:"data_source_fields"`
	EventoccurrenceAsWinner2Urls []string  `json:"eventoccurrence_as_winner_2_urls"`
	DataSourceID                 string    `json:"data_source_id"`
	DriveroccurrenceUrls         []string  `json:"driveroccurrence_urls"`
	ImageUrls                    []string  `json:"image_urls"`
	LastDataIngest               time.Time `json:"last_data_ingest"`
	EventoccurrenceAsWinner3Urls []string  `json:"eventoccurrence_as_winner_3_urls"`
	Language                     string    `json:"language"`
	Created                      time.Time `json:"created"`
	Modified                     time.Time `json:"modified"`
	ContentUrls                  []string  `json:"content_urls"`
	TeamURL                      string    `json:"team_url"`
	Editability                  string    `json:"editability"`
	DriverRacingnumber           int       `json:"driver_racingnumber"`
}

type teamStruct struct {
	Name                 string    `json:"name"`
	Language             string    `json:"language"`
	Created              time.Time `json:"created"`
	Colour               string    `json:"colour"`
	DriveroccurrenceUrls []string  `json:"driveroccurrence_urls"`
	DriverUrls           []string  `json:"driver_urls"`
	Modified             time.Time `json:"modified"`
	ImageUrls            []string  `json:"image_urls"`
	NationURL            string    `json:"nation_url"`
	ContentUrls          []string  `json:"content_urls"`
	LastDataIngest       time.Time `json:"last_data_ingest"`
	DataSourceFields     []string  `json:"data_source_fields"`
	Self                 string    `json:"self"`
	Editability          string    `json:"editability"`
	DataSourceID         string    `json:"data_source_id"`
	UID                  string    `json:"uid"`
}

type seasonStruct struct {
	Name                     string        `json:"name"`
	Language                 string        `json:"language"`
	Created                  time.Time     `json:"created"`
	ScheduleUrls             []string      `json:"schedule_urls"`
	Self                     string        `json:"self"`
	HasContent               bool          `json:"has_content"`
	ImageUrls                []string      `json:"image_urls"`
	Modified                 time.Time     `json:"modified"`
	ScheduleAfterNextYearURL string        `json:"schedule_after_next_year_url"`
	LastDataIngest           time.Time     `json:"last_data_ingest"`
	DataSourceFields         []interface{} `json:"data_source_fields"`
	Year                     int           `json:"year"`
	EventoccurrenceUrls      []string      `json:"eventoccurrence_urls"`
	Editability              string        `json:"editability"`
	DataSourceID             string        `json:"data_source_id"`
	UID                      string        `json:"uid"`
}

type allSeasonStruct struct {
	Seasons []seasonStruct `json:"objects"`
}

type eventStruct struct {
	EventURL              string    `json:"event_url"`
	UID                   string    `json:"uid"`
	RaceSeasonURL         string    `json:"race_season_url"`
	ScheduleUrls          []string  `json:"schedule_urls"`
	Winner3URL            string    `json:"winner_3_url"`
	OfficialName          string    `json:"official_name"`
	NationURL             string    `json:"nation_url"`
	SessionoccurrenceUrls []string  `json:"sessionoccurrence_urls"`
	CircuitURL            string    `json:"circuit_url"`
	Self                  string    `json:"self"`
	DataSourceFields      []string  `json:"data_source_fields"`
	StartDate             string    `json:"start_date"`
	DataSourceID          string    `json:"data_source_id"`
	EndDate               string    `json:"end_date"`
	ImageUrls             []string  `json:"image_urls"`
	Slug                  string    `json:"slug"`
	LastDataIngest        time.Time `json:"last_data_ingest"`
	Winner2URL            string    `json:"winner_2_url"`
	Name                  string    `json:"name"`
	Language              string    `json:"language"`
	Created               time.Time `json:"created"`
	Modified              time.Time `json:"modified"`
	SponsorURL            string    `json:"sponsor_url"`
	Winner1URL            string    `json:"winner_1_url"`
	Editability           string    `json:"editability"`
}

type sessionStruct struct {
	UID                      string        `json:"uid"`
	ScheduleAfterMidnightURL string        `json:"schedule_after_midnight_url"`
	ScheduleUrls             []string      `json:"schedule_urls"`
	SessionExpiredTime       time.Time     `json:"session_expired_time"`
	ChannelUrls              []string      `json:"channel_urls"`
	GlobalChannelUrls        []string      `json:"global_channel_urls"`
	AvailableForUser         bool          `json:"available_for_user"`
	ScheduleAfter7DaysURL    string        `json:"schedule_after_7_days_url"`
	NbcStatus                string        `json:"nbc_status"`
	Self                     string        `json:"self"`
	ReplayStartTime          time.Time     `json:"replay_start_time"`
	DataSourceFields         []interface{} `json:"data_source_fields"`
	DataSourceID             string        `json:"data_source_id"`
	Status                   string        `json:"status"`
	ScheduleAfter14DaysURL   string        `json:"schedule_after_14_days_url"`
	EventoccurrenceURL       string        `json:"eventoccurrence_url"`
	DriveroccurrenceUrls     []interface{} `json:"driveroccurrence_urls"`
	StartTime                time.Time     `json:"start_time"`
	ImageUrls                []string      `json:"image_urls"`
	LiveSourcesPath          string        `json:"live_sources_path"`
	StatusOverride           interface{}   `json:"status_override"`
	NbcPid                   int           `json:"nbc_pid"`
	LiveSourcesMd5           string        `json:"live_sources_md5"`
	Slug                     string        `json:"slug"`
	LastDataIngest           time.Time     `json:"last_data_ingest"`
	Name                     string        `json:"name"`
	SessionTypeURL           string        `json:"session_type_url"`
	EditorialStartTime       time.Time     `json:"editorial_start_time"`
	EventConfigMd5           string        `json:"event_config_md5"`
	EditorialEndTime         interface{}   `json:"editorial_end_time"`
	Language                 string        `json:"language"`
	Created                  time.Time     `json:"created"`
	Modified                 time.Time     `json:"modified"`
	ContentUrls              []string      `json:"content_urls"`
	ScheduleAfter24HURL      string        `json:"schedule_after_24h_url"`
	EndTime                  time.Time     `json:"end_time"`
	SeriesURL                string        `json:"series_url"`
	SessionName              string        `json:"session_name"`
	Editability              string        `json:"editability"`
}

type sessionStreamsStruct struct {
	Objects []struct {
		Status         string `json:"status"`
		SessionTypeURL struct {
			Name string `json:"name"`
		} `json:"session_type_url"`
		EditorialStartTime time.Time `json:"editorial_start_time"`
		NbcStatus          string    `json:"nbc_status"`
		EventoccurrenceURL struct {
			CircuitURL struct {
				ShortName string `json:"short_name"`
			} `json:"circuit_url"`
			Slug string `json:"slug"`
		} `json:"eventoccurrence_url"`
		LiveSourcesPath   string              `json:"live_sources_path"`
		UID               string              `json:"uid"`
		ChannelUrls       []channelUrlsStruct `json:"channel_urls"`
		GlobalChannelUrls []struct {
			Self string `json:"self"`
			Slug string `json:"slug"`
			UID  string `json:"uid"`
		} `json:"global_channel_urls"`
		DataSourceID     string `json:"data_source_id"`
		AvailableForUser bool   `json:"available_for_user"`
	} `json:"objects"`
}

type channelUrlsStruct struct {
	UID        string             `json:"uid"`
	Self       string             `json:"self"`
	DriverUrls []driverUrlsStruct `json:"driver_urls"`
	Ovps       []struct {
		AccountURL    string `json:"account_url"`
		Path          string `json:"path"`
		Domain        string `json:"domain"`
		FullStreamURL string `json:"full_stream_url"`
	} `json:"ovps"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type driverUrlsStruct struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ImageUrls []struct {
		ImageType string `json:"image_type"`
		URL       string `json:"url"`
	} `json:"image_urls"`
	DriverTla string `json:"driver_tla"`
	TeamURL   struct {
		Colour string `json:"colour"`
		Name   string `json:"name"`
	} `json:"team_url"`
	DriverRacingnumber int `json:"driver_racingnumber"`
}
type homepageContentStruct struct {
	Objects []struct {
		Items []struct {
			Position   int `json:"position"`
			ContentURL struct {
				Items []struct {
					Position    int    `json:"position"`
					ContentType string `json:"content_type"`
					ContentURL  struct {
						Self string `json:"self"`
						UID  string `json:"uid"`
					} `json:"content_url"`
				} `json:"items"`
				Self        string `json:"self"`
				UID         string `json:"uid"`
				SetTypeSlug string `json:"set_type_slug"`
				Title       string `json:"title"`
			} `json:"content_url"`
			ContentType string `json:"content_type"`
			DisplayType string `json:"display_type,omitempty"`
		} `json:"items"`
		Slug        string `json:"slug"`
		SetTypeSlug string `json:"set_type_slug"`
	} `json:"objects"`
}

//downloads json from URL and returns the json as string and whether it's valid as bool
func getJSON(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	response := buf.String()
	return response, nil
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func getDriver(driverID string) driverStruct {
	var driver driverStruct
	jsonString, err := getJSON(urlStart + driverID)
	if err != nil {
		return driver
	}
	json.Unmarshal([]byte(jsonString), &driver)
	return driver
}

func getTeam(teamID string) teamStruct {
	var team teamStruct
	jsonString, err := getJSON(urlStart + teamID)
	if err != nil {
		return team
	}
	json.Unmarshal([]byte(jsonString), &team)
	return team
}

func getEpisode(episodeID string) episodeStruct {
	var ep episodeStruct
	jsonString, err := getJSON(urlStart + episodeID)
	if err != nil {
		return ep
	}
	json.Unmarshal([]byte(jsonString), &ep)
	return ep
}

func getHomepageContent() homepageContentStruct {
	var home homepageContentStruct
	jsonString, err := getJSON("https://f1tv.formula1.com/api/sets/?slug=home&fields=slug,set_type_slug,items,items__position,items__content_type,items__display_type,items__content_url,items__content_url__uid,items__content_url__self,items__content_url__set_type_slug,items__content_url__display_type_slug,items__content_url__title,items__content_url__items,items__content_url__items__set_type_slug,items__content_url__items__position,items__content_url__items__content_type,items__content_url__items__content_url,items__content_url__items__content_url__self,items__content_url__items__content_url__uid&fields_to_expand=items__content_url,items__content_url__items__content_url")
	if err != nil {
		return home
	}
	json.Unmarshal([]byte(jsonString), &home)
	return home
}

func getVodTypes() vodTypesStruct {
	var types vodTypesStruct
	jsonString, err := getJSON(vodTypesURL)
	if err != nil {
		return types
	}
	json.Unmarshal([]byte(jsonString), &types)
	return types
}

var listOfSeasons allSeasonStruct

func getSeasons() allSeasonStruct {
	if len(listOfSeasons.Seasons) < 1 {
		jsonString, err := getJSON("https://f1tv.formula1.com/api/race-season/?fields=year,name,self,has_content,eventoccurrence_urls&year__gt=2017&order=year")
		if err != nil {
			return listOfSeasons
		}
		json.Unmarshal([]byte(jsonString), &listOfSeasons)
	}
	return listOfSeasons
}

func getEvent(eventID string) eventStruct {
	var event eventStruct
	jsonString, err := getJSON(urlStart + eventID)
	if err != nil {
		return event
	}
	json.Unmarshal([]byte(jsonString), &event)
	return event
}

func getSession(sessionID string) sessionStruct {
	var session sessionStruct
	jsonString, err := getJSON(urlStart + sessionID)
	if err != nil {
		return session
	}
	json.Unmarshal([]byte(jsonString), &session)
	return session
}

func getSessionStreams(sessionSlug string) sessionStreamsStruct {
	var sessionStreams sessionStreamsStruct
	jsonString, err := getJSON(sessionURLstart + sessionSlug)
	if err != nil {
		return sessionStreams
	}
	json.Unmarshal([]byte(jsonString), &sessionStreams)
	return sessionStreams
}
