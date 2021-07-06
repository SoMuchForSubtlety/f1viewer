package f1tv

type APIResponse struct {
	ResultCode       string    `json:"resultCode,omitempty"`
	Message          string    `json:"message,omitempty"`
	ErrorDescription string    `json:"errorDescription,omitempty"`
	ResultObj        ResultObj `json:"resultObj,omitempty"`
	SystemTime       int64     `json:"systemTime,omitempty"`
	Source           string    `json:"source,omitempty"`
}

type Category struct {
	ExternalPathIds []string `json:"externalPathIds,omitempty"`
	StartDate       int64    `json:"startDate,omitempty"`
	CategoryID      int      `json:"categoryId,omitempty"`
	EndDate         int64    `json:"endDate,omitempty"`
	CategoryPathIds []int    `json:"categoryPathIds,omitempty"`
	OrderID         int      `json:"orderId,omitempty"`
	IsPrimary       bool     `json:"isPrimary,omitempty"`
	CategoryName    string   `json:"categoryName,omitempty"`
}

type Bundles struct {
	BundleID      int    `json:"bundleId,omitempty"`
	BundleType    string `json:"bundleType,omitempty"`
	BundleSubtype string `json:"bundleSubtype,omitempty"`
	IsParent      bool   `json:"isParent,omitempty"`
	OrderID       int    `json:"orderId,omitempty"`
}

type TechnicalPackage struct {
	PackageID   int    `json:"packageId,omitempty"`
	PackageName string `json:"packageName,omitempty"`
	PackageType string `json:"packageType,omitempty"`
}

type PlatformVariants struct {
	SubtitlesLanguages []interface{}      `json:"subtitlesLanguages,omitempty"`
	AudioLanguages     []interface{}      `json:"audioLanguages,omitempty"`
	TechnicalPackages  []TechnicalPackage `json:"technicalPackages,omitempty"`
	CpID               int                `json:"cpId,omitempty"`
	VideoType          string             `json:"videoType,omitempty"`
	PictureURL         string             `json:"pictureUrl,omitempty"`
	TrailerURL         string             `json:"trailerUrl,omitempty"`
	HasTrailer         bool               `json:"hasTrailer,omitempty"`
}

type Properties struct {
	MeetingNumber        int    `json:"meeting_Number,omitempty"`
	SessionEndTime       int64  `json:"sessionEndTime,omitempty"`
	Series               string `json:"series,omitempty"`
	LastUpdatedDate      int64  `json:"lastUpdatedDate,omitempty"`
	SeasonMeetingOrdinal int    `json:"season_Meeting_Ordinal,omitempty"`
	MeetingStartDate     int64  `json:"meeting_Start_Date,omitempty"`
	MeetingEndDate       int64  `json:"meeting_End_Date,omitempty"`
	Season               int    `json:"season,omitempty"`
	SessionIndex         int    `json:"session_index,omitempty"`
	SessionStartDate     int64  `json:"sessionStartDate,omitempty"`
	MeetingSessionKey    int    `json:"meetingSessionKey,omitempty"`
	SessionEndDate       int64  `json:"sessionEndDate,omitempty"`
}

type EmfAttributes struct {
	VideoType                      string `json:"VideoType,omitempty"`
	MeetingKey                     string `json:"MeetingKey,omitempty"`
	MeetingSessionKey              string `json:"MeetingSessionKey,omitempty"`
	MeetingName                    string `json:"Meeting_Name,omitempty"`
	MeetingNumber                  string `json:"Meeting_Number,omitempty"`
	CircuitShortName               string `json:"Circuit_Short_Name,omitempty"`
	MeetingCode                    string `json:"Meeting_Code,omitempty"`
	MeetingCountryKey              string `json:"MeetingCountryKey,omitempty"`
	CircuitKey                     string `json:"CircuitKey,omitempty"`
	MeetingLocation                string `json:"Meeting_Location,omitempty"`
	Series                         string `json:"Series,omitempty"`
	OBC                            bool   `json:"OBC,omitempty"`
	State                          string `json:"state,omitempty"`
	TimetableKey                   string `json:"TimetableKey,omitempty"`
	SessionKey                     string `json:"SessionKey,omitempty"`
	SessionPeriod                  string `json:"SessionPeriod,omitempty"`
	CircuitOfficialName            string `json:"Circuit_Official_Name,omitempty"`
	ActivityDescription            string `json:"ActivityDescription,omitempty"`
	SeriesMeetingSessionIdentifier string `json:"SeriesMeetingSessionIdentifier,omitempty"`
	SessionEndTime                 string `json:"sessionEndTime,omitempty"`
	MeetingStartDate               string `json:"Meeting_Start_Date,omitempty"`
	MeetingEndDate                 string `json:"Meeting_End_Date,omitempty"`
	TrackLength                    string `json:"Track_Length,omitempty"`
	ScheduledLapCount              string `json:"Scheduled_Lap_Count,omitempty"`
	ScheduledDistance              string `json:"Scheduled_Distance,omitempty"`
	CircuitLocation                string `json:"Circuit_Location,omitempty"`
	MeetingSponsor                 string `json:"Meeting_Sponsor,omitempty"`
	IsTestEvent                    string `json:"IsTestEvent,omitempty"`
	ChampionshipMeetingOrdinal     string `json:"Championship_Meeting_Ordinal,omitempty"`
	MeetingOfficialName            string `json:"Meeting_Official_Name,omitempty"`
	MeetingDisplayDate             string `json:"Meeting_Display_Date,omitempty"`
	PageID                         PageID `json:"PageID,omitempty"`
	MeetingCountryName             string `json:"Meeting_Country_Name,omitempty"`
	GlobalTitle                    string `json:"Global_Title,omitempty"`
	GlobalMeetingCountryName       string `json:"Global_Meeting_Country_Name,omitempty"`
	GlobalMeetingName              string `json:"Global_Meeting_Name,omitempty"`
	DriversID                      string `json:"Drivers_ID,omitempty"`
	Year                           string `json:"Year,omitempty"`
	TeamsID                        string `json:"Teams_ID,omitempty"`
	// inconsistent types
	// SeasonMeetingOrdinal           int         `json:"Season_Meeting_Ordinal,omitempty"`
	// SessionStartDate               int64       `json:"sessionStartDate,omitempty"`
	// SessionEndDate                 int64       `json:"sessionEndDate,omitempty"`
	// SessionIndex                   int         `json:"session_index,omitempty"`
}

type Language []struct {
	LanguageCode string `json:"languageCode,omitempty"`
	LanguageName string `json:"languageName,omitempty"`
}

type Metadata struct {
	EmfAttributes      EmfAttributes        `json:"emfAttributes,omitempty"`
	LongDescription    string               `json:"longDescription,omitempty"`
	Country            string               `json:"country,omitempty"`
	Year               string               `json:"year,omitempty"`
	ContractStartDate  int64                `json:"contractStartDate,omitempty"`
	EpisodeNumber      int                  `json:"episodeNumber,omitempty"`
	ContractEndDate    int64                `json:"contractEndDate,omitempty"`
	ExternalID         string               `json:"externalId,omitempty"`
	Title              string               `json:"title,omitempty"`
	TitleBrief         string               `json:"titleBrief,omitempty"`
	ObjectType         string               `json:"objectType,omitempty"`
	Duration           int64                `json:"duration,omitempty"`
	Genres             []string             `json:"genres,omitempty"`
	ContentSubtype     ContentSubType       `json:"contentSubtype,omitempty"`
	PcLevel            int                  `json:"pcLevel,omitempty"`
	ContentID          int64                `json:"contentId,omitempty"`
	StarRating         int                  `json:"starRating,omitempty"`
	PictureURL         string               `json:"pictureUrl,omitempty"`
	ContentType        ContentType          `json:"contentType,omitempty"`
	Language           string               `json:"language,omitempty"`
	Season             int                  `json:"season,omitempty"`
	UIDuration         string               `json:"uiDuration,omitempty"`
	Entitlement        string               `json:"entitlement,omitempty"`
	Locked             bool                 `json:"locked,omitempty"`
	Label              string               `json:"label,omitempty"`
	ImageURL           string               `json:"imageUrl,omitempty"`
	ID                 string               `json:"id,omitempty"`
	MetaDescription    string               `json:"meta-description,omitempty"`
	IsADVAllowed       bool                 `json:"isADVAllowed,omitempty"`
	ContentProvider    string               `json:"contentProvider,omitempty"`
	IsLatest           bool                 `json:"isLatest,omitempty"`
	IsOnAir            bool                 `json:"isOnAir,omitempty"`
	IsEncrypted        bool                 `json:"isEncrypted,omitempty"`
	ObjectSubtype      string               `json:"objectSubtype,omitempty"`
	MetadataLanguage   string               `json:"metadataLanguage,omitempty"`
	PcLevelVod         string               `json:"pcLevelVod,omitempty"`
	IsParent           bool                 `json:"isParent,omitempty"`
	AvailableLanguages []AvailableLanguages `json:"availableLanguages,omitempty"`
	AdvTags            string               `json:"advTags,omitempty"`
	ShortDescription   string               `json:"shortDescription,omitempty"`
	LeavingSoon        bool                 `json:"leavingSoon,omitempty"`
	AvailableAlso      []string             `json:"availableAlso,omitempty"`
	PcVodLabel         string               `json:"pcVodLabel,omitempty"`
	IsGeoBlocked       bool                 `json:"isGeoBlocked,omitempty"`
	Filter             string               `json:"filter,omitempty"`
	ComingSoon         bool                 `json:"comingSoon,omitempty"`
	IsPopularEpisode   bool                 `json:"isPopularEpisode,omitempty"`
	PrimaryCategoryID  int                  `json:"primaryCategoryId,omitempty"`
	MeetingKey         string               `json:"meetingKey,omitempty"`
	VideoType          string               `json:"videoType,omitempty"`
	ParentalAdvisory   string               `json:"parentalAdvisory,omitempty"`
	AdditionalStreams  []AdditionalStream   `json:"additionalStreams,omitempty"`
}

type Container struct {
	ID               string             `json:"id,omitempty"`
	Layout           string             `json:"layout,omitempty"`
	Actions          []Actions          `json:"actions,omitempty"`
	PlatformVariants []PlatformVariants `json:"platformVariants,omitempty"`
	Properties       []Properties       `json:"properties,omitempty"`
	Metadata         Metadata           `json:"metadata,omitempty"`
	RetrieveItems    RetrieveItems      `json:"retrieveItems,omitempty"`
	Translations     Translations       `json:"translations,omitempty"`
	Categories       []Category         `json:"categories,omitempty"`
	Bundles          []Bundles          `json:"bundles,omitempty"`
}

type ContentContainer struct {
	ID               string             `json:"id,omitempty"`
	Layout           string             `json:"layout,omitempty"`
	Actions          []Actions          `json:"actions,omitempty"`
	PlatformVariants []PlatformVariants `json:"platformVariants,omitempty"`
	Properties       []Properties       `json:"properties,omitempty"`
	Metadata         Metadata           `json:"metadata,omitempty"`
	Containers       struct {
		Categories []Category `json:"categories,omitempty"`
		Bundles    []Bundles  `json:"bundles,omitempty"`
	} `json:"containers,omitempty"`
}

type ContentDetailsContainer struct{}

type TopContainer struct {
	// inconsistent type
	// ID            string        `json:"id,omitempty"`
	Layout        string        `json:"layout,omitempty"`
	Actions       []Actions     `json:"actions,omitempty"`
	Metadata      Metadata      `json:"metadata,omitempty"`
	RetrieveItems RetrieveItems `json:"retrieveItems,omitempty"`
	Translations  Translations  `json:"translations,omitempty"`

	// only in content details
	PlatformVariants []PlatformVariants `json:"platformVariants,omitempty"`
	ContentID        int64              `json:"contentId,omitempty"`
	Containers       struct {
		Bundles    []Bundles    `json:"bundles,omitempty"`
		Categories []Categories `json:"categories,omitempty"`
	} `json:"containers,omitempty"`
	Suggest      Suggest      `json:"suggest,omitempty"`
	PlatformName string       `json:"platformName,omitempty"`
	Properties   []Properties `json:"properties,omitempty"`
}

type ResultObj struct {
	Total       int            `json:"total,omitempty"`
	Containers  []TopContainer `json:"containers,omitempty"`
	MeetingName string         `json:"meetingName,omitempty"`
	Metadata    Metadata       `json:"metadata,omitempty"`
}

type ContainerResultObj struct {
	Total       int                `json:"total,omitempty"`
	Containers  []ContentContainer `json:"containers,omitempty"`
	MeetingName string             `json:"meetingName,omitempty"`
	Metadata    Metadata           `json:"metadata,omitempty"`
}

type RetrieveItems struct {
	ResultObj    ContainerResultObj `json:"resultObj,omitempty"`
	URIOriginal  string             `json:"uriOriginal,omitempty"`
	TypeOriginal string             `json:"typeOriginal,omitempty"`
}

type Actions struct {
	Key        string `json:"key,omitempty"`
	URI        string `json:"uri,omitempty"`
	TargetType string `json:"targetType,omitempty"`
	Type       string `json:"type,omitempty"`
	Layout     string `json:"layout,omitempty"`
	HREF       string `json:"href,omitempty"`
}

type MetadataLabel struct {
	NLD string `json:"NLD,omitempty"`
	FRA string `json:"FRA,omitempty"`
	DEU string `json:"DEU,omitempty"`
	POR string `json:"POR,omitempty"`
	SPA string `json:"SPA,omitempty"`
}

type Translations struct {
	MetadataLabel MetadataLabel `json:"metadata.label,omitempty"`
}

type AvailableLanguages struct {
	LanguageCode string `json:"languageCode,omitempty"`
	LanguageName string `json:"languageName,omitempty"`
}

type AdditionalStream struct {
	RacingNumber    int    `json:"racingNumber,omitempty"`
	Title           string `json:"title,omitempty"`
	DriverFirstName string `json:"driverFirstName,omitempty"`
	DriverLastName  string `json:"driverLastName,omitempty"`
	TeamName        string `json:"teamName,omitempty"`
	ConstructorName string `json:"constructorName,omitempty"`
	Type            string `json:"type,omitempty"`
	PlaybackURL     string `json:"playbackUrl,omitempty"`
	DriverImg       string `json:"driverImg,omitempty"`
	TeamImg         string `json:"teamImg,omitempty"`
	Hex             string `json:"hex,omitempty"`
}

type AudioLanguages struct {
	AudioLanguageName string `json:"audioLanguageName,omitempty"`
	AudioID           string `json:"audioId,omitempty"`
	IsPreferred       bool   `json:"isPreferred,omitempty"`
}

type TechnicalPackages struct {
	PackageID   int    `json:"packageId,omitempty"`
	PackageName string `json:"packageName,omitempty"`
	PackageType string `json:"packageType,omitempty"`
}

type Categories struct {
	CategoryPathIds []int    `json:"categoryPathIds,omitempty"`
	ExternalPathIds []string `json:"externalPathIds,omitempty"`
	EndDate         int64    `json:"endDate,omitempty"`
	OrderID         int      `json:"orderId,omitempty"`
	IsPrimary       bool     `json:"isPrimary,omitempty"`
	CategoryName    string   `json:"categoryName,omitempty"`
	CategoryID      int      `json:"categoryId,omitempty"`
	StartDate       int64    `json:"startDate,omitempty"`
}

type Containers struct{}

type Suggest struct {
	Input   []string `json:"input,omitempty"`
	Payload struct {
		ObjectSubtype string `json:"objectSubtype,omitempty"`
		ContentID     string `json:"contentId,omitempty"`
		Title         string `json:"title,omitempty"`
		ObjectType    string `json:"objectType,omitempty"`
	} `json:"payload,omitempty"`
}
