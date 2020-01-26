package main

import (
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/SoMuchForSubtlety/golark"
)

var endpoint = "https://f1tv.formula1.com/api/"

type episode struct {
	Title        string   `json:"title"`
	Subtitle     string   `json:"subtitle"`
	UID          string   `json:"uid"`
	DataSourceID string   `json:"data_source_id"`
	Items        []string `json:"items"`
}

type vodTypes struct {
	Objects []struct {
		Name        string   `json:"name"`
		ContentUrls []string `json:"content_urls"`
		UID         string   `json:"uid"`
	} `json:"objects"`
}

type team struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`
	UID    string `json:"uid"`
}

type seasonStruct struct {
	Name                string   `json:"name"`
	HasContent          bool     `json:"has_content"`
	Year                int      `json:"year"`
	EventoccurrenceUrls []string `json:"eventoccurrence_urls"`
	UID                 string   `json:"uid"`
}

type seasons struct {
	Seasons []seasonStruct `json:"objects"`
}

type eventStruct struct {
	UID                   string   `json:"uid"`
	Name                  string   `json:"name"`
	OfficialName          string   `json:"official_name"`
	SessionoccurrenceUrls []string `json:"sessionoccurrence_urls"`
	StartDate             string   `json:"start_date"`
	EndDate               string   `json:"end_date"`
}

type sessionStruct struct {
	UID         string    `json:"uid"`
	SessionName string    `json:"session_name"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	ContentUrls []string  `json:"content_urls"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

type channel struct {
	UID        string   `json:"uid"`
	Self       string   `json:"self"`
	Name       string   `json:"name"`
	DriverUrls []driver `json:"driver_urls"`
}

type driver struct {
	LastName           string `json:"last_name"`
	UID                string `json:"uid"`
	FirstName          string `json:"first_name"`
	DriverTla          string `json:"driver_tla"`
	TeamURL            team   `json:"team_url"`
	DriverRacingnumber int    `json:"driver_racingnumber"`
}

type collectionItem struct {
	Archived    bool   `json:"archived"`
	UID         string `json:"uid"`
	Language    string `json:"language"`
	ContentType string `json:"content_type"`
	ContentURL  string `json:"content_url"`
	DisplayType string `json:"display_type,omitempty"`
	SetURL      string `json:"set_url"`
}

type collection struct {
	UID         string           `json:"uid"`
	Title       string           `json:"title"`
	UniqueItems bool             `json:"unique_items"`
	Items       []collectionItem `json:"items"`
	Summary     string           `json:"summary"`
}

type collectionList struct {
	Objects []collection `json:"objects"`
}

func getCollectionList() (collList collectionList, err error) {
	err = golark.NewRequest(endpoint, "sets", "").
		AddField(golark.NewField("title")).
		AddField(golark.NewField("uid")).
		WithFilter("set_type_slug", golark.NewFilter(golark.Equals, "video")).
		Execute(&collList)
	return
}

func getCollection(collID string) (coll collection, err error) {
	err = golark.NewRequest(endpoint, "sets", collID).
		AddField(golark.NewField("items")).
		Execute(&coll)
	return
}

func getHomepageContent() (collection, error) {
	type container struct {
		Objects []collection `json:"objects"`
	}

	var response container
	err := golark.NewRequest(endpoint, "sets", "").
		AddField(golark.NewField("items")).
		WithFilter("slug", golark.NewFilter(golark.Equals, "home")).
		Execute(&response)

	if len(response.Objects) == 0 {
		return collection{}, err
	}
	return response.Objects[0], err
}

func getVodTypes() (types vodTypes, err error) {
	err = golark.NewRequest(endpoint, "vod-type-tag", "").
		AddField(golark.NewField("name")).
		AddField(golark.NewField("content_urls")).
		Execute(&types)
	return
}

func getSeasons() (s seasons, err error) {
	year := golark.NewField("year").WithFilter(golark.NewFilter(golark.GreaterThan, "2017"))
	err = golark.NewRequest(endpoint, "race-season", "").
		AddField(year).
		AddField(golark.NewField("name")).
		AddField(golark.NewField("has_content")).
		AddField(golark.NewField("eventoccurrence_urls")).
		OrderBy(year).
		Execute(&s)
	return
}

func getEvent(eventID string) (event eventStruct, err error) {
	// TODO: use proper ID
	err = golark.NewRequest(endpoint, "event-occurrence", pathToUID(eventID)).
		AddField(golark.NewField("name")).
		AddField(golark.NewField("sessionoccurrence_urls")).
		Execute(&event)
	return
}

func getSession(sessionID string) (session sessionStruct, err error) {
	err = golark.NewRequest(endpoint, "session-occurrence", pathToUID(sessionID)).
		AddField(golark.NewField("name")).
		AddField(golark.NewField("status")).
		AddField(golark.NewField("uid")).
		Execute(&session)
	return
}

func getSessions(sessionIDs []string) ([]sessionStruct, error) {
	type container struct {
		Objects []sessionStruct `json:"objects"`
	}

	var response container

	for i, id := range sessionIDs {
		sessionIDs[i] = pathToUID(id)
	}

	err := golark.NewRequest(endpoint, "session-occurrence", "").
		AddField(golark.NewField("name")).
		AddField(golark.NewField("status")).
		AddField(golark.NewField("content_urls")).
		AddField(golark.NewField("uid").
			WithFilter(golark.NewFilter(golark.Equals, strings.Join(sessionIDs, ",")))).
		Execute(&response)

	return response.Objects, err
}

func getSessionStreams(sessionID string) ([]channel, error) {
	type container struct {
		Channels []channel `json:"channel_urls"`
	}
	var channels container

	err := golark.NewRequest(endpoint, "session-occurrence", sessionID).
		AddField(golark.NewField("channel_urls").
			WithSubField(golark.NewField("self")).
			WithSubField(golark.NewField("name")).
			WithSubField(golark.NewField("driver_urls").
				WithSubField(golark.NewField("driver_racingnumber")).
				WithSubField(golark.NewField("team_url").
					WithSubField(golark.NewField("name")).
					WithSubField(golark.NewField("colour"))))).
		Execute(&channels)

	return channels.Channels, err
}

func (s *viewerSession) loadEpisodes(episodeIDs []string) ([]episode, error) {
	type container struct {
		Objects []episode `json:"objects"`
	}

	const batchSize = 5
	for i, id := range episodeIDs {
		episodeIDs[i] = pathToUID(id)
	}

	episodes := make([]episode, len(episodeIDs))

	var wg sync.WaitGroup
	wg.Add(len(episodes) / batchSize)
	if len(episodeIDs)%batchSize > 0 {
		wg.Add(1)
	}
	for i := 0; i < len(episodeIDs); i += batchSize {
		go func(rangeStart int) {
			rangeEnd := rangeStart + batchSize
			if rangeEnd > len(episodeIDs) {
				rangeEnd = len(episodeIDs)
			}

			query := strings.Join(episodeIDs[rangeStart:rangeEnd], ",")
			var response container
			// TODO: properly handle error
			err := golark.NewRequest(endpoint, "episodes", "").
				AddField(golark.NewField("title")).
				AddField(golark.NewField("subtitle")).
				AddField(golark.NewField("uid").
					WithFilter(golark.NewFilter(golark.Equals, query))).
				AddField(golark.NewField("data_source_id")).
				AddField(golark.NewField("items")).
				Execute(&response)
			if err != nil {
				s.logError(err)
				return
			}
			copy(episodes[rangeStart:], response.Objects)
			wg.Done()
		}(i)
	}
	wg.Wait()
	return episodes, nil
}

func sortEpisodes(episodes []episode) []episode {
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

func pathToUID(p string) (uid string) {
	return path.Base(p)
}
