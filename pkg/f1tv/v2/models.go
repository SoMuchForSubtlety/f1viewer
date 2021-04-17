package f1tv

type APIResponse struct {
	ResultCode       string    `json:"resultCode"`
	Message          string    `json:"message"`
	ErrorDescription string    `json:"errorDescription"`
	ResultObj        ResultObj `json:"resultObj"`
	SystemTime       int       `json:"systemTime"`
	Source           string    `json:"source"`
}

type Category struct {
	ExternalPathIds []string `json:"externalPathIds"`
	StartDate       int64    `json:"startDate"`
	CategoryID      int      `json:"categoryId"`
	EndDate         int64    `json:"endDate"`
	CategoryPathIds []int    `json:"categoryPathIds"`
	OrderID         int      `json:"orderId"`
	IsPrimary       bool     `json:"isPrimary"`
	CategoryName    string   `json:"categoryName"`
}

type Bundles struct {
	BundleID      int    `json:"bundleId"`
	BundleType    string `json:"bundleType"`
	BundleSubtype string `json:"bundleSubtype"`
	IsParent      bool   `json:"isParent"`
	OrderID       int    `json:"orderId"`
}

type TechnicalPackage struct {
	PackageID   int    `json:"packageId"`
	PackageName string `json:"packageName"`
	PackageType string `json:"packageType"`
}

type PlatformVariants struct {
	SubtitlesLanguages []interface{}      `json:"subtitlesLanguages"`
	AudioLanguages     []interface{}      `json:"audioLanguages"`
	TechnicalPackages  []TechnicalPackage `json:"technicalPackages"`
}

type Properties struct {
	MeetingNumber        int    `json:"meeting_Number"`
	SessionEndTime       int64  `json:"sessionEndTime"`
	Series               string `json:"series"`
	LastUpdatedDate      int64  `json:"lastUpdatedDate"`
	SeasonMeetingOrdinal int    `json:"season_Meeting_Ordinal"`
	MeetingStartDate     int    `json:"meeting_Start_Date"`
	MeetingEndDate       int    `json:"meeting_End_Date"`
	Season               int    `json:"season"`
	SessionIndex         int    `json:"session_index"`
	SessionStartDate     int    `json:"sessionStartDate"`
	MeetingSessionKey    int    `json:"meetingSessionKey"`
	SessionEndDate       int    `json:"sessionEndDate"`
}

type EmfAttributes struct {
	VideoType                      string      `json:"VideoType"`
	MeetingKey                     string      `json:"MeetingKey"`
	MeetingSessionKey              string      `json:"MeetingSessionKey"`
	MeetingName                    string      `json:"Meeting_Name"`
	MeetingNumber                  string      `json:"Meeting_Number"`
	CircuitShortName               string      `json:"Circuit_Short_Name"`
	MeetingCode                    string      `json:"Meeting_Code"`
	MeetingCountryKey              string      `json:"MeetingCountryKey"`
	CircuitKey                     string      `json:"CircuitKey"`
	MeetingLocation                string      `json:"Meeting_Location"`
	Series                         string      `json:"Series"`
	OBC                            bool        `json:"OBC"`
	State                          string      `json:"state"`
	TimetableKey                   string      `json:"TimetableKey"`
	SessionKey                     string      `json:"SessionKey"`
	SessionPeriod                  string      `json:"SessionPeriod"`
	CircuitOfficialName            string      `json:"Circuit_Official_Name"`
	ActivityDescription            string      `json:"ActivityDescription"`
	SeriesMeetingSessionIdentifier string      `json:"SeriesMeetingSessionIdentifier"`
	SessionEndTime                 string      `json:"sessionEndTime"` // TODO: time
	MeetingStartDate               string      `json:"Meeting_Start_Date"`
	MeetingEndDate                 string      `json:"Meeting_End_Date"`
	TrackLength                    string      `json:"Track_Length"`
	ScheduledLapCount              string      `json:"Scheduled_Lap_Count"`
	ScheduledDistance              string      `json:"Scheduled_Distance"`
	CircuitLocation                string      `json:"Circuit_Location"`
	MeetingSponsor                 string      `json:"Meeting_Sponsor"`
	IsTestEvent                    string      `json:"IsTestEvent"`
	SeasonMeetingOrdinal           int         `json:"Season_Meeting_Ordinal"`
	ChampionshipMeetingOrdinal     string      `json:"Championship_Meeting_Ordinal"`
	SessionIndex                   int         `json:"session_index"`
	MeetingOfficialName            string      `json:"Meeting_Official_Name"`
	MeetingDisplayDate             string      `json:"Meeting_Display_Date"`
	PageID                         interface{} `json:"PageID"`
	MeetingCountryName             string      `json:"Meeting_Country_Name"`
	SessionEndDate                 int64       `json:"sessionEndDate"`
	SessionStartDate               int64       `json:"sessionStartDate"`
	GlobalTitle                    string      `json:"Global_Title"`
	GlobalMeetingCountryName       string      `json:"Global_Meeting_Country_Name"`
	GlobalMeetingName              string      `json:"Global_Meeting_Name"`
	DriversID                      string      `json:"Drivers_ID"`
	Year                           string      `json:"Year"`
	TeamsID                        string      `json:"Teams_ID"`
}

type Language []struct {
	LanguageCode string `json:"languageCode"`
	LanguageName string `json:"languageName"`
}

type Metadata struct {
	EmfAttributes     EmfAttributes  `json:"emfAttributes"`
	LongDescription   string         `json:"longDescription"`
	Country           string         `json:"country"`
	Year              string         `json:"year"`
	ContractStartDate int64          `json:"contractStartDate"`
	EpisodeNumber     int            `json:"episodeNumber"`
	ContractEndDate   int64          `json:"contractEndDate"`
	ExternalID        string         `json:"externalId"`
	AvailableAlso     []string       `json:"availableAlso"`
	Title             string         `json:"title"`
	TitleBrief        string         `json:"titleBrief"`
	ObjectType        string         `json:"objectType"`
	Duration          int            `json:"duration"`
	Genres            []string       `json:"genres"`
	ContentSubtype    ContentSubType `json:"contentSubtype"`
	PcLevel           int            `json:"pcLevel"`
	ContentID         int            `json:"contentId"`
	StarRating        int            `json:"starRating"`
	PictureURL        string         `json:"pictureUrl"`
	ContentType       ContentType    `json:"contentType"`
	Language          string         `json:"language"`
	Season            int            `json:"season"`
	UIDuration        string         `json:"uiDuration"`
	Entitlement       string         `json:"entitlement"`
	Locked            bool           `json:"locked"`
	Label             string         `json:"label"`
	ImageURL          string         `json:"imageUrl"`
	ID                string         `json:"id"`
	MetaDescription   string         `json:"meta-description"`
}

type Container struct {
	ID               string             `json:"id"`
	Layout           string             `json:"layout"`
	Actions          []Actions          `json:"actions"`
	PlatformVariants []PlatformVariants `json:"platformVariants"`
	Properties       []Properties       `json:"properties"`
	Metadata         Metadata           `json:"metadata"`
	RetrieveItems    RetrieveItems      `json:"retrieveItems"`
	Translations     Translations       `json:"translations,omitempty"`
	Categories       []Category         `json:"categories"`
	Bundles          []Bundles          `json:"bundles"`
}

type ContentContainer struct {
	ID               string             `json:"id"`
	Layout           string             `json:"layout"`
	Actions          []Actions          `json:"actions"`
	PlatformVariants []PlatformVariants `json:"platformVariants"`
	Properties       []Properties       `json:"properties"`
	Metadata         Metadata           `json:"metadata"`
	Containers       struct {
		Categories []Category `json:"categories"`
		Bundles    []Bundles  `json:"bundles"`
	} `json:"containers"`
}

type TopContainer struct {
	ID            string        `json:"id"`
	Layout        string        `json:"layout"`
	Actions       []Actions     `json:"actions"`
	Metadata      Metadata      `json:"metadata"`
	RetrieveItems RetrieveItems `json:"retrieveItems"`
	Translations  Translations  `json:"translations,omitempty"`
}

type ResultObj struct {
	Total       int            `json:"total"`
	Containers  []TopContainer `json:"containers"`
	MeetingName string         `json:"meetingName"`
	Metadata    Metadata       `json:"metadata"`
}

type ContainerResultObj struct {
	Total       int                `json:"total"`
	Containers  []ContentContainer `json:"containers"`
	MeetingName string             `json:"meetingName"`
	Metadata    Metadata           `json:"metadata"`
}

type RetrieveItems struct {
	ResultObj    ContainerResultObj `json:"resultObj"`
	URIOriginal  string             `json:"uriOriginal"`
	TypeOriginal string             `json:"typeOriginal"`
}

type Actions struct {
	Key        string `json:"key"`
	URI        string `json:"uri"`
	TargetType string `json:"targetType"`
	Type       string `json:"type"`
	Layout     string `json:"layout"`
}

type MetadataLabel struct {
	NLD string `json:"NLD"`
	FRA string `json:"FRA"`
	DEU string `json:"DEU"`
	POR string `json:"POR"`
	SPA string `json:"SPA"`
}

type Translations struct {
	MetadataLabel MetadataLabel `json:"metadata.label"`
}
