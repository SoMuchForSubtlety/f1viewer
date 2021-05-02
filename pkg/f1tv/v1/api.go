package f1tv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
	"github.com/SoMuchForSubtlety/golark"
	"golang.org/x/sync/errgroup"
)

var endpoint = "https://f1tv.formula1.com/api/"

const (
	liveSlug = "grand-prix-weekend-live"
)

var headers = http.Header{
	// "User-Agent": []string{},
}

type F1TV struct {
	AuthToken string
	plan      SubscriptionPlan

	userAgent string
	Client    http.Client
}

func NewF1TV(version string) *F1TV {
	return &F1TV{
		Client:    *http.DefaultClient,
		userAgent: fmt.Sprintf("f1viewer/%s (%s)", version, runtime.GOOS),
	}
}

type Episode struct {
	Title        string   `json:"title"`
	Subtitle     string   `json:"subtitle"`
	UID          string   `json:"uid"`
	DataSourceID string   `json:"data_source_id"`
	Items        []string `json:"items"`
}

type VodTypes struct {
	Objects []struct {
		Name        string   `json:"name"`
		ContentUrls []string `json:"content_urls"`
		UID         string   `json:"uid"`
	} `json:"objects"`
}

type Season struct {
	Name                string   `json:"name"`
	HasContent          bool     `json:"has_content"`
	Year                int      `json:"year"`
	EventoccurrenceUrls []string `json:"eventoccurrence_urls"`
	UID                 string   `json:"uid"`
}

type Event struct {
	UID                   string   `json:"uid"`
	Name                  string   `json:"name"`
	OfficialName          string   `json:"official_name"`
	SessionoccurrenceUrls []string `json:"sessionoccurrence_urls"`
	EndDate               ISODate  `json:"end_date"`
}

type Session struct {
	UID         string    `json:"uid"`
	SessionName string    `json:"session_name"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	ContentUrls []string  `json:"content_urls"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

type Channel struct {
	UID  string `json:"uid"`
	Self string `json:"self"`
	Name string `json:"name"`
}

type Plan struct {
	Amount        int      `json:"amount"`
	ContentURL    string   `json:"content_url"`
	Currency      string   `json:"currency"`
	ID            int      `json:"id"`
	PricePointURL string   `json:"price_point_url"`
	Self          string   `json:"self"`
	CsgItemUrls   []string `json:"csg_item_urls"`
	UID           string   `json:"uid"`
	DataSourceID  string   `json:"data_source_id"`
	ObjectID      int      `json:"object_id"`
	Name          string   `json:"name"`
	Recurring     bool     `json:"recurring"`
	Interval      string   `json:"interval"`
	IntervalCount int      `json:"interval_count"`
	Sku           string   `json:"sku"`
	StripeID      int      `json:"stripe_id"`
	Product       Product  `json:"product"`
}

type Product struct {
	Type string `json:"type"`
	Slug string `json:"slug"`
}

func (c Channel) PrettyName() string {
	switch c.Name {
	case "WIF":
		return "Main Feed"
	case "pit lane":
		return "Pit Lane"
	case "driver":
		return "Driver Tracker"
	case "data":
		return "Data Channel"
	default:
		return c.Name
	}
}

type CollectionItem struct {
	Archived    bool   `json:"archived"`
	UID         string `json:"uid"`
	Language    string `json:"language"`
	ContentType string `json:"content_type"`
	ContentURL  string `json:"content_url"`
	DisplayType string `json:"display_type,omitempty"`
	SetURL      string `json:"set_url"`
}

type Collection struct {
	UID         string           `json:"uid"`
	Title       string           `json:"title"`
	UniqueItems bool             `json:"unique_items"`
	Items       []CollectionItem `json:"items"`
	Summary     string           `json:"summary"`
}

type ISODate struct {
	Format string
	time.Time
}

func (Date *ISODate) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	Date.Format = "2006-01-02"
	t, _ := time.Parse(Date.Format, s)
	Date.Time = t
	return nil
}

func (Date ISODate) MarshalJSON() ([]byte, error) {
	return json.Marshal(Date.Time.Format(Date.Format))
}

func GetPlan(uri string) (Plan, error) {
	var plan Plan
	err := golark.NewRequest(endpoint, "plans", path.Base(uri)).
		Headers(headers).
		Execute(&plan)

	return plan, err
}

func GetLiveWeekendEvent() (Event, bool, error) {
	type container struct {
		Objects []Collection `json:"objects"`
	}

	var liveSet container
	err := golark.NewRequest(endpoint, "sets", "").
		AddField(golark.NewField("items")).
		WithFilter("slug", golark.NewFilter(golark.Equals, liveSlug)).
		Headers(headers).
		Execute(&liveSet)
	if err != nil {
		return Event{}, false, err
	}
	if len(liveSet.Objects) == 0 || len(liveSet.Objects[0].Items) == 0 {
		return Event{}, false, nil
	}
	event, err := GetEvent(liveSet.Objects[0].Items[0].ContentURL)
	return event, true, err
}

func GetCollectionList() ([]Collection, error) {
	var collList struct {
		Objects []Collection `json:"objects"`
	}

	err := golark.NewRequest(endpoint, "sets", "").
		AddField(golark.NewField("title")).
		AddField(golark.NewField("uid")).
		WithFilter("set_type_slug", golark.NewFilter(golark.Equals, "video")).
		Headers(headers).
		Execute(&collList)
	return collList.Objects, err
}

func GetCollection(collID string) (coll Collection, err error) {
	err = golark.NewRequest(endpoint, "sets", collID).
		AddField(golark.NewField("items")).
		Headers(headers).
		Execute(&coll)
	return
}

func GetVodTypes() (types VodTypes, err error) {
	err = golark.NewRequest(endpoint, "vod-type-tag", "").
		AddField(golark.NewField("name")).
		AddField(golark.NewField("content_urls")).
		Headers(headers).
		Execute(&types)
	return
}

func GetSeasons() ([]Season, error) {
	year := golark.NewField("year")
	var s struct {
		Seasons []Season `json:"objects"`
	}

	err := golark.NewRequest(endpoint, "race-season", "").
		AddField(golark.NewField("year")).
		AddField(golark.NewField("name")).
		AddField(golark.NewField("has_content")).
		AddField(golark.NewField("eventoccurrence_urls")).
		OrderBy(year, golark.Descending).
		Headers(headers).
		Execute(&s)

	return s.Seasons, err
}

func GetEvent(eventID string) (event Event, err error) {
	err = golark.NewRequest(endpoint, "event-occurrence", pathToUID(eventID)).
		AddField(golark.NewField("name")).
		AddField(golark.NewField("end_date")).
		AddField(golark.NewField("sessionoccurrence_urls")).
		Headers(headers).
		Execute(&event)
	return
}

func GetSession(sessionID string) (session Session, err error) {
	err = golark.NewRequest(endpoint, "session-occurrence", pathToUID(sessionID)).
		AddField(golark.NewField("name")).
		AddField(golark.NewField("status")).
		AddField(golark.NewField("uid")).
		AddField(golark.NewField("session_name")).
		AddField(golark.NewField("start_time")).
		AddField(golark.NewField("end_time")).
		Headers(headers).
		Execute(&session)
	return
}

func GetSessions(sessionIDs []string) ([]Session, error) {
	type container struct {
		Objects []Session `json:"objects"`
	}

	var response container

	for i, id := range sessionIDs {
		sessionIDs[i] = pathToUID(id)
	}

	err := golark.NewRequest(endpoint, "session-occurrence", "").
		AddField(golark.NewField("name")).
		AddField(golark.NewField("status")).
		AddField(golark.NewField("content_urls")).
		AddField(golark.NewField("start_time")).
		AddField(golark.NewField("end_time")).
		AddField(golark.NewField("uid").
			WithFilter(golark.NewFilter(golark.Equals, strings.Join(sessionIDs, ",")))).
		Headers(headers).
		Execute(&response)

	return response.Objects, err
}

func GetSessionStreams(sessionID string) ([]Channel, error) {
	type container struct {
		Channels []Channel `json:"channel_urls"`
	}
	var channels container

	err := golark.NewRequest(endpoint, "session-occurrence", sessionID).
		AddField(golark.NewField("channel_urls").
			WithSubField(golark.NewField("self")).
			WithSubField(golark.NewField("name")).
			WithSubField(golark.NewField("uid"))).
		Headers(headers).
		Execute(&channels)

	return channels.Channels, err
}

func LoadEpisodes(episodeIDs []string) ([]Episode, error) {
	type container struct {
		Objects []Episode `json:"objects"`
	}

	const batchSize = 5
	for i, id := range episodeIDs {
		episodeIDs[i] = pathToUID(id)
	}

	episodes := make([]Episode, len(episodeIDs))

	var loadingGroup errgroup.Group
	for i := 0; i < len(episodeIDs); i += batchSize {
		rangeStart := i
		loadingGroup.Go(func() error {
			rangeEnd := rangeStart + batchSize
			if rangeEnd > len(episodeIDs) {
				rangeEnd = len(episodeIDs)
			}

			query := strings.Join(episodeIDs[rangeStart:rangeEnd], ",")
			var response container
			err := golark.NewRequest(endpoint, "episodes", "").
				AddField(golark.NewField("title")).
				AddField(golark.NewField("subtitle")).
				AddField(golark.NewField("uid").
					WithFilter(golark.NewFilter(golark.Equals, query))).
				AddField(golark.NewField("data_source_id")).
				AddField(golark.NewField("items")).
				Headers(headers).
				Execute(&response)
			if err != nil {
				return err
			}
			copy(episodes[rangeStart:], response.Objects)
			return nil
		})
	}
	err := loadingGroup.Wait()
	return sortEpisodes(episodes), err
}

func sortEpisodes(episodes []Episode) []Episode {
	sort.Slice(episodes, func(i, j int) bool {
		if len(episodes[i].DataSourceID) >= 4 && len(episodes[j].DataSourceID) >= 4 {
			year1, race1, err := util.GetYearAndRace(episodes[i].DataSourceID)
			year2, race2, err2 := util.GetYearAndRace(episodes[j].DataSourceID)
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
