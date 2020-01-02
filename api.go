package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

const urlStart = "https://f1tv.formula1.com"
const sessionURLstart = "https://f1tv.formula1.com/api/session-occurrence/?fields=uid,nbc_status,status,editorial_start_time,live_sources_path,data_source_id,available_for_user,global_channel_urls,global_channel_urls__uid,global_channel_urls__slug,global_channel_urls__self,channel_urls,channel_urls__ovps,channel_urls__slug,channel_urls__name,channel_urls__uid,channel_urls__self,channel_urls__driver_urls,channel_urls__driver_urls__driver_tla,channel_urls__driver_urls__driver_racingnumber,channel_urls__driver_urls__first_name,channel_urls__driver_urls__last_name,channel_urls__driver_urls__image_urls,channel_urls__driver_urls__image_urls__image_type,channel_urls__driver_urls__image_urls__url,channel_urls__driver_urls__team_url,channel_urls__driver_urls__team_url__name,channel_urls__driver_urls__team_url__colour,eventoccurrence_url,eventoccurrence_url__slug,eventoccurrence_url__circuit_url,eventoccurrence_url__circuit_url__short_name,session_type_url,session_type_url__name&fields_to_expand=global_channel_urls,channel_urls,channel_urls__driver_urls,channel_urls__driver_urls__image_urls,channel_urls__driver_urls__team_url,eventoccurrence_url,eventoccurrence_url__circuit_url,session_type_url&slug="
const homepageContentURL = "https://f1tv.formula1.com/api/sets/?slug=home&fields=items"
const seasonsSince2017URL = "https://f1tv.formula1.com/api/race-season/?fields=year,name,self,has_content,eventoccurrence_urls&year__gt=2017&order=year"
const f2chasingCollID = "/api/sets/coll_4440e712d31d42fb95c9a2145ab4dac7/"

const tagsURL = "https://f1tv.formula1.com/api/tags/"
const vodTypesURL = "http://f1tv.formula1.com/api/vod-type-tag/"
const seriesListURL = "https://f1tv.formula1.com/api/series/"
const seriesF1URL = "https://f1tv.formula1.com/api/series/seri_436bb431c3a24d7d8e200a74e1d11de4/"
const teamsURL = "https://f1tv.formula1.com/api/episodes/"
const collListURL = "https://f1tv.formula1.com/api/sets/?fields=title,self&set_type_slug=video"

type episodeStruct struct {
	Subtitle              string        `json:"subtitle"`
	UID                   string        `json:"uid"`
	ScheduleUrls          []string      `json:"schedule_urls"`
	SessionoccurrenceUrls []string      `json:"sessionoccurrence_urls"`
	Stats                 interface{}   `json:"stats"`
	Title                 string        `json:"title"`
	TierUrls              []string      `json:"tier_urls"`
	Self                  string        `json:"self"`
	DriverUrls            []string      `json:"driver_urls"`
	CircuitUrls           []interface{} `json:"circuit_urls"`
	ProgramEntryMd5       string        `json:"program_entry_md5"`
	DataSourceFields      []string      `json:"data_source_fields"`
	ParentURL             interface{}   `json:"parent_url"`
	DataSourceID          string        `json:"data_source_id"`
	VodTypeTagUrls        []string      `json:"vod_type_tag_urls"`
	Tags                  []struct {
		ScheduleUrls []string `json:"schedule_urls"`
		TagURL       string   `json:"tag_url"`
	} `json:"tags"`
	ImageUrls              []string      `json:"image_urls"`
	SeriesUrls             []string      `json:"series_urls"`
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
	Name                     string        `json:"name"`
	Slug                     string        `json:"slug"`
	Status                   string        `json:"status"`
	ContentUrls              []string      `json:"content_urls"`
	SessionName              string        `json:"session_name"`
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
	ScheduleAfter14DaysURL   string        `json:"schedule_after_14_days_url"`
	EventoccurrenceURL       string        `json:"eventoccurrence_url"`
	DriveroccurrenceUrls     []interface{} `json:"driveroccurrence_urls"`
	StartTime                time.Time     `json:"start_time"`
	ImageUrls                []string      `json:"image_urls"`
	LiveSourcesPath          string        `json:"live_sources_path"`
	StatusOverride           interface{}   `json:"status_override"`
	NbcPid                   int           `json:"nbc_pid"`
	LiveSourcesMd5           string        `json:"live_sources_md5"`
	LastDataIngest           time.Time     `json:"last_data_ingest"`
	SessionTypeURL           string        `json:"session_type_url"`
	EditorialStartTime       time.Time     `json:"editorial_start_time"`
	EventConfigMd5           string        `json:"event_config_md5"`
	EditorialEndTime         interface{}   `json:"editorial_end_time"`
	Language                 string        `json:"language"`
	Created                  time.Time     `json:"created"`
	Modified                 time.Time     `json:"modified"`
	ScheduleAfter24HURL      string        `json:"schedule_after_24h_url"`
	EndTime                  time.Time     `json:"end_time"`
	SeriesURL                string        `json:"series_url"`
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
type homepageContent struct {
	Objects []struct {
		Items []struct {
			Archived              bool          `json:"archived"`
			UID                   string        `json:"uid"`
			Language              string        `json:"language"`
			Created               time.Time     `json:"created"`
			ScheduleUrls          []string      `json:"schedule_urls"`
			Self                  string        `json:"self"`
			Modified              time.Time     `json:"modified"`
			ImageUrls             []interface{} `json:"image_urls"`
			ContentType           string        `json:"content_type"`
			ContentURL            string        `json:"content_url"`
			DisplayType           string        `json:"display_type,omitempty"`
			DataSourceFields      []interface{} `json:"data_source_fields"`
			DataSourceID          interface{}   `json:"data_source_id"`
			Position              int           `json:"position"`
			ScheduledItemModified time.Time     `json:"scheduled_item_modified"`
			SetURL                string        `json:"set_url"`
			Editability           string        `json:"editability"`
			LastDataIngest        interface{}   `json:"last_data_ingest"`
			TextReview            string        `json:"text_review,omitempty"`
			SubCollection         bool          `json:"sub_collection,omitempty"`
		} `json:"items"`
	} `json:"objects"`
}

type collection struct {
	UID              string        `json:"uid"`
	ScheduleUrls     []string      `json:"schedule_urls"`
	Stats            interface{}   `json:"stats"`
	Title            string        `json:"title"`
	UniqueItems      bool          `json:"unique_items"`
	Self             string        `json:"self"`
	DataSourceFields []interface{} `json:"data_source_fields"`
	HasPrice         bool          `json:"has_price"`
	SetTypeURL       string        `json:"set_type_url"`
	DataSourceID     interface{}   `json:"data_source_id"`
	Body             string        `json:"body"`
	Plans            []interface{} `json:"plans"`
	Tags             []interface{} `json:"tags"`
	ImageUrls        []interface{} `json:"image_urls"`
	HierarchyURL     interface{}   `json:"hierarchy_url"`
	SponsorUrls      []interface{} `json:"sponsor_urls"`
	Slug             string        `json:"slug"`
	LastDataIngest   interface{}   `json:"last_data_ingest"`
	Language         string        `json:"language"`
	Created          time.Time     `json:"created"`
	Items            []struct {
		Archived              bool          `json:"archived"`
		UID                   string        `json:"uid"`
		Language              string        `json:"language"`
		Created               time.Time     `json:"created"`
		ScheduleUrls          []string      `json:"schedule_urls"`
		Self                  string        `json:"self"`
		Modified              time.Time     `json:"modified"`
		ImageUrls             []interface{} `json:"image_urls"`
		ContentType           string        `json:"content_type"`
		ContentURL            string        `json:"content_url"`
		DisplayType           interface{}   `json:"display_type"`
		DataSourceFields      []interface{} `json:"data_source_fields"`
		DataSourceID          interface{}   `json:"data_source_id"`
		Position              int           `json:"position"`
		ScheduledItemModified time.Time     `json:"scheduled_item_modified"`
		SetURL                string        `json:"set_url"`
		Editability           string        `json:"editability"`
		LastDataIngest        interface{}   `json:"last_data_ingest"`
		TextReview            string        `json:"text_review"`
	} `json:"items"`
	Modified    time.Time `json:"modified"`
	Summary     string    `json:"summary"`
	SetTypeSlug string    `json:"set_type_slug"`
	Editability string    `json:"editability"`
}

type collectionList struct {
	Objects []collection `json:"objects"`
}

func getCollectionList() (collectionList, error) {
	var collList collectionList
	err := doGet(collListURL, &collList)
	return collList, err
}

func getCollection(collID string) (collection, error) {
	var coll collection
	err := doGet(urlStart+collID, &coll)
	return coll, err
}

func getDriver(driverID string) (driverStruct, error) {
	var driver driverStruct
	err := doGet(urlStart+driverID, &driver)
	return driver, err
}

func getTeam(teamID string) (teamStruct, error) {
	var team teamStruct
	err := doGet(urlStart+teamID, &team)
	return team, err
}

func getEpisode(episodeID string) (episodeStruct, error) {
	var ep episodeStruct
	err := doGet(urlStart+episodeID, &ep)
	return ep, err
}

func getHomepageContent() (homepageContent, error) {
	var home homepageContent
	err := doGet(homepageContentURL, &home)
	return home, err
}

func getVodTypes() (vodTypesStruct, error) {
	var types vodTypesStruct
	err := doGet(vodTypesURL, &types)
	return types, err
}

var listOfSeasons allSeasonStruct

func getSeasons() (allSeasonStruct, error) {
	var err error
	if len(listOfSeasons.Seasons) < 1 {
		err = doGet(seasonsSince2017URL, &listOfSeasons)
	}
	return listOfSeasons, err
}

func getEvent(eventID string) (eventStruct, error) {
	var event eventStruct
	err := doGet(urlStart+eventID, &event)
	return event, err
}

func getSession(sessionID string) (sessionStruct, error) {
	var session sessionStruct
	err := doGet(urlStart+sessionID, &session)
	return session, err
}

func getSessionStreams(sessionSlug string) (sessionStreamsStruct, error) {
	var sessionStreams sessionStreamsStruct
	err := doGet(sessionURLstart+sessionSlug, &sessionStreams)
	return sessionStreams, err
}

func (s *viewerSession) loadEpisodes(IDs []string) ([]episodeStruct, error) {
	var episodes []episodeStruct
	errChan := make(chan error)
	// TODO: tweak number of threads
	guard := make(chan struct{}, 100)
	var er error
	for i := range IDs {
		// wait for space in guard
		guard <- struct{}{}
		go func(i int) {
			epID := IDs[i]
			// check if episode metadata is already cached
			s.episodeMapMutex.RLock()
			ep, ok := s.episodeMap[epID]
			s.episodeMapMutex.RUnlock()
			if !ok {
				// load episode metadata and add to cache
				var err error
				ep, err = getEpisode(epID)
				if err != nil {
					errChan <- err
					return
				}
				s.episodeMapMutex.Lock()
				s.episodeMap[epID] = ep
				s.episodeMapMutex.Unlock()
			}
			// maybe not thread safe
			episodes = append(episodes, ep)
			// make room in guard
			<-guard
			errChan <- nil
		}(i)
	}
	for index := 0; index < len(IDs); index++ {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		}
	}
	return episodes, er
}

func sortEpisodes(episodes []episodeStruct) []episodeStruct {
	sort.Slice(episodes, func(i, j int) bool {
		if len(episodes[i].DataSourceID) >= 4 && len(episodes[j].DataSourceID) >= 4 {
			year1, race1, err := getYearAndRace(episodes[i].DataSourceID)
			year2, race2, err2 := getYearAndRace(episodes[j].DataSourceID)
			if err == nil && err2 == nil {
				// sort chronologically by year and race number
				if year1 != year2 {
					return year1 < year2
				} else if race1 != race2 {
					return race1 < race2
				}
			}
		}
		return episodes[i].Title < episodes[j].Title
	})
	return episodes
}

func doGet(url string, v interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyBytes, v)
}
