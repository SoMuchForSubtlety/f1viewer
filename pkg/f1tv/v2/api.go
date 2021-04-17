package f1tv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	baseURL             = "https://f1tv.formula1.com"
	backupStreamURL     = "https://f1tv.formula1.com/dr/stream.json"
	authURL             = "https://api.formula1.com/v2/account/subscriber/authenticate/by-password"
	pathStart           = "/2.0/R/ENG/%v/ALL/"
	playbackRequestPath = "/CONTENT/PLAY?contentId=%v"

	apiKey = "fCUCjWrKPu9ylJwRAv8BpGLEgiAuThx7"

	BIG_SCREEN_HLS  StreamType = "BIG_SCREEN_HLS"
	WEB_HLS         StreamType = "WEB_HLS"
	TABLET_HLS      StreamType = "TABLET_HLS"
	MOBILE_HLS      StreamType = "MOBILE_HLS"
	BIG_SCREEN_DASH StreamType = "BIG_SCREEN_DASH"
	WEB_DASH        StreamType = "WEB_DASH"
	MOBILE_DASH     StreamType = "MOBILE_DASH"
	TABLET_DASH     StreamType = "TABLET_DASH"

	CATEGORY_LIVE RequestCategory = 395

	VIDEO ContentType = "VIDEO"

	LIVE   ContentSubType = "LIVE"
	REPLAY ContentSubType = "REPLAY"
)

type ContentType string

type ContentSubType string

type StreamType string

type RequestCategory int

type F1TV struct {
	SubscriptionToken string
	userAgent         string
	Client            *http.Client
}

func NewF1TV(version string) *F1TV {
	return &F1TV{
		userAgent: fmt.Sprintf("f1viewer/%s (%s)", version, runtime.GOOS),
		Client:    http.DefaultClient,
	}
}

func (f *F1TV) Authenticate(username, password string) error {
	type request struct {
		Login    string `json:"Login"`
		Password string `json:"Password"`
	}

	payloadBuf := new(bytes.Buffer)
	err := json.NewEncoder(payloadBuf).Encode(request{Login: username, Password: password})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, authURL, payloadBuf)
	req.Header.Set("apiKey", apiKey)
	req.Header.Set("User-Agent", "RaceControl f1viewer")
	if err != nil {
		return err
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return err
	}
	var auth struct {
		Data struct {
			SubscriptionStatus string `json:"subscriptionStatus"`
			SubscriptionToken  string `json:"subscriptionToken"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&auth)
	f.SubscriptionToken = auth.Data.SubscriptionToken
	return err
}

func GetContent(format StreamType, category RequestCategory, v interface{}) error {
	resp, err := http.Get(fmt.Sprintf(baseURL+filepath.Join(pathStart, "PAGE/%v/F1_TV_Pro_Annual/2"), format, category))
	if err != nil {
		return fmt.Errorf("error during request: %w", err)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (f *F1TV) GetVideoContainers() ([]TopContainer, error) {
	var resp APIResponse
	err := GetContent(WEB_DASH, CATEGORY_LIVE, &resp)
	if err != nil {
		return nil, err
	}

	var nonEmpty []TopContainer
	for _, container := range resp.ResultObj.Containers {
		var videoContainers []ContentContainer
		for _, contentContainer := range container.RetrieveItems.ResultObj.Containers {
			if contentContainer.Metadata.ContentType == VIDEO {
				videoContainers = append(videoContainers, contentContainer)
			}
		}
		container.RetrieveItems.ResultObj.Containers = videoContainers
		if len(videoContainers) > 0 {
			nonEmpty = append(nonEmpty, container)
		}
	}

	return nonEmpty, err
}

func (f *F1TV) GetLiveVideoContainers() ([]ContentContainer, error) {
	topContainers, err := f.GetVideoContainers()
	if err != nil {
		return nil, err
	}
	var live []ContentContainer
	for _, vidContainers := range topContainers {
		for _, v := range vidContainers.RetrieveItems.ResultObj.Containers {
			if v.Metadata.ContentSubtype == LIVE {
				live = append(live, v)
			}
		}
	}

	return live, nil
}

func (f *F1TV) GetPlaybackURL(format StreamType, contentID int) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(baseURL+"/1.0/R/ENG/%v/ALL/CONTENT/PLAY", format), nil)
	if err != nil {
		return "", nil
	}

	q := req.URL.Query()
	q.Add("contentId", strconv.Itoa(contentID))
	req.URL.RawQuery = q.Encode()
	req.Header.Set("ascendontoken", f.SubscriptionToken)
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var resp struct {
		ResultCode       string `json:"resultCode"`
		Message          string `json:"message"`
		ErrorDescription string `json:"errorDescription"`
		ResultObj        struct {
			EntitlementToken string `json:"entitlementToken"`
			URL              string `json:"url"`
			StreamType       string `json:"streamType"`
		} `json:"resultObj"`
		SystemTime int64 `json:"systemTime"`
	}

	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return "", err
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		err = errors.New(resp.Message)
	}

	return resp.ResultObj.URL, err
}
