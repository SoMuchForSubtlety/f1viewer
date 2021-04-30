package f1tv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
)

const (
	baseURL = "https://f1tv.formula1.com"
	authURL = "https://api.formula1.com/v2/account/subscriber/authenticate/by-password"

	playbackRequestPath            = "/1.0/R/ENG/%v/ALL/CONTENT/PLAY?contentId=%d"
	playbackPerspectiveRequestPath = "/1.0/R/ENG/%v/ALL/%s"
	contentDetailsPath             = "/2.0/R/ENG/%v/ALL/CONTENT/VIDEO/%d/F1_TV_Pro_Annual/14"
	categoryPagePath               = "/2.0/R/ENG/%v/ALL/PAGE/%v/F1_TV_Pro_Annual/2"

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

func assembleURL(urlPath string, format StreamType, args ...interface{}) (*url.URL, error) {
	args = append([]interface{}{format}, args...)
	return url.Parse(baseURL + fmt.Sprintf(urlPath, args...))
}

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

func (f *F1TV) GetContent(format StreamType, category RequestCategory, v interface{}) error {
	reqURL, err := assembleURL(categoryPagePath, format, category)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return err
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error during request: %w", err)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (f *F1TV) GetVideoContainers() ([]TopContainer, error) {
	var resp APIResponse
	err := f.GetContent(WEB_DASH, CATEGORY_LIVE, &resp)
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
	ids := make(map[int]struct{})
	for _, vidContainers := range topContainers {
		for _, v := range vidContainers.RetrieveItems.ResultObj.Containers {
			_, ok := ids[v.Metadata.ContentID]
			if !ok && v.Metadata.ContentSubtype == LIVE {
				ids[v.Metadata.ContentID] = struct{}{}
				live = append(live, v)
			}
		}
	}

	return live, nil
}

func (f *F1TV) ContentDetails(contentID int) (TopContainer, error) {
	reqURL, err := assembleURL(contentDetailsPath, BIG_SCREEN_HLS, contentID)
	if err != nil {
		return TopContainer{}, err
	}
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return TopContainer{}, err
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return TopContainer{}, err
	}
	defer resp.Body.Close()

	var details APIResponse
	err = json.NewDecoder(resp.Body).Decode(&details)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TopContainer{}, fmt.Errorf("got status code %d: %s", resp.StatusCode, details.Message)
	}

	if len(details.ResultObj.Containers) == 0 {
		return TopContainer{}, fmt.Errorf("no content details for %d", contentID)
	}
	return details.ResultObj.Containers[0], err
}

func (f *F1TV) GetPerspectivePlaybackURL(format StreamType, path string) (string, error) {
	reqURL, err := assembleURL(playbackPerspectiveRequestPath, format, path)
	if err != nil {
		return "", nil
	}

	return f.playbackURL(reqURL.String())
}

func (f *F1TV) GetPlaybackURL(format StreamType, contentID int) (string, error) {
	reqURL, err := assembleURL(playbackRequestPath, format, contentID)
	if err != nil {
		return "", nil
	}

	return f.playbackURL(reqURL.String())
}

func (f *F1TV) playbackURL(reqURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", nil
	}

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
