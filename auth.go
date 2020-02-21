package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	identityProvider = "/api/identity-providers/iden_732298a17f9c458890a1877880d140f3/"
	authURL          = "https://api.formula1.com/v2/account/subscriber/authenticate/by-password"
	getTokenURL      = "https://f1tv-api.formula1.com/agl/1.0/unk/en/all_devices/global/authenticate"
	apiKey           = "fCUCjWrKPu9ylJwRAv8BpGLEgiAuThx7"
)

type authResponse struct {
	Data struct {
		SubscriptionStatus string `json:"subscriptionStatus"`
		SubscriptionToken  string `json:"subscriptionToken"`
	} `json:"data"`
}

type tokenResponse struct {
	PlanUrls          []string `json:"plan_urls"`
	Token             string   `json:"token"`
	UserIsVip         bool     `json:"user_is_vip"`
	Oauth2AccessToken string   `json:"oauth2_access_token"`
}

func (session *viewerSession) login() (string, error) {
	auth, err := authenticate(session.cfg.Username, session.cfg.Password)
	if err != nil {
		return "", fmt.Errorf("could not log in: %w", err)
	}
	token, err := getToken(auth.Data.SubscriptionToken)
	if err != nil {
		return "", fmt.Errorf("could not authenticate in: %w", err)
	}
	return token.Token, nil
}

func authenticate(username, password string) (authResponse, error) {
	type request struct {
		Login    string `json:"Login"`
		Password string `json:"Password"`
	}

	header := http.Header{}
	header["apiKey"] = []string{apiKey}
	respBody, err := post(request{Login: username, Password: password}, authURL, header)
	if err != nil {
		return authResponse{}, err
	}

	var auth authResponse
	err = json.Unmarshal(respBody, &auth)
	return auth, err
}

func getToken(accessToken string) (tokenResponse, error) {
	type request struct {
		IdentityProviderURL string `json:"identity_provider_url"`
		AccessToken         string `json:"access_token"`
	}

	// TODO: double ckeck auth providers
	respBody, err := post(request{IdentityProviderURL: identityProvider, AccessToken: accessToken}, getTokenURL, http.Header{})
	if err != nil {
		return tokenResponse{}, err
	}

	var token tokenResponse
	err = json.Unmarshal(respBody, &token)
	return token, err
}

func post(content interface{}, url string, header http.Header) ([]byte, error) {
	body, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = checkResponse(resp)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respString, err := ioutil.ReadAll(resp.Body)
		if err != nil || string(respString) == "" {
			return fmt.Errorf("got status %s", resp.Status)
		}
		return fmt.Errorf("got status %s with body:\n%s", resp.Status, respString)
	}
	return nil
}

func (session *viewerSession) testAuth() {
	token, err := session.login()
	if err != nil {
		session.logError("could not log in ", err)
	} else {
		session.authtoken = token
		session.logInfo("login successful!")
	}
}
