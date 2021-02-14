package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SoMuchForSubtlety/keyring"
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
	if session.authtoken != "" {
		if tokenValid(session.authtoken) {
			return session.authtoken, nil
		}
	}

	auth, err := authenticate(session.username, session.password)
	if err != nil {
		return "", fmt.Errorf("could not log in: %w", err)
	}
	token, err := session.getToken(auth.Data.SubscriptionToken)
	if err != nil {
		return "", fmt.Errorf("could not authenticate in: %w", err)
	}
	return token.Token, nil
}

func tokenValid(token string) bool {
	_, err := getPlayableURL("/api/channels/chan_d77f90b2775f4db4855d32605f2c65da/", token)
	return err == nil
}

func (session *viewerSession) logout() {
	err := session.removeCredentials()
	if err != nil {
		session.logError("Failed to log out:", err)
	} else {
		session.logInfo("logged out!")
	}
	session.authtoken = ""
}

func authenticate(username, password string) (authResponse, error) {
	type request struct {
		Login    string `json:"Login"`
		Password string `json:"Password"`
	}

	header := http.Header{}
	header.Set("apiKey", apiKey)
	header.Set("User-Agent", "RaceControl f1viewer")
	respBody, err := post(request{Login: username, Password: password}, authURL, header)
	if err != nil {
		return authResponse{}, err
	}

	var auth authResponse
	err = json.Unmarshal(respBody, &auth)
	return auth, err
}

func (session *viewerSession) getToken(accessToken string) (tokenResponse, error) {
	type request struct {
		IdentityProviderURL string `json:"identity_provider_url"`
		AccessToken         string `json:"access_token"`
	}

	// TODO: double ckeck auth providers
	respBody, err := post(request{identityProvider, accessToken}, getTokenURL, headers)
	if err != nil {
		return tokenResponse{}, err
	}

	var token tokenResponse
	err = json.Unmarshal(respBody, &token)
	go session.checkPlans(token)
	return token, err
}

func (session *viewerSession) checkPlans(token tokenResponse) {
	if len(token.PlanUrls) == 0 {
		session.logInfo("looks like you don't have an F1TV subscription, some streams might not be accessible")
		return
	}

	plan, err := getPlan(token.PlanUrls[0])
	if err != nil {
		session.logErrorf("failed to get user subscription information: %v", err)
		return
	}

	switch plan.Product.Slug {
	case "pro":
		session.logInfo("detected active F1TV pro subscription")
	case "access":
		session.logInfo("looks like you have an F1TV access subscription, some streams might not be accessible")
	default:
		session.logInfof("unknown F1TV subscription tier '%s'", plan.Product.Slug)
	}
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
	if session.authtoken != "" {
		_, err := getPlayableURL("/api/channels/chan_d77f90b2775f4db4855d32605f2c65da/", session.authtoken)
		if err != nil {
			session.logError(err)
		} else {
			session.logInfo("token works!")
		}
	} else {
		token, err := session.login()
		if err != nil {
			session.logError(err)
		} else {
			session.authtoken = token
			session.logInfo("login successful!")
		}
	}
}

func (session *viewerSession) openRing() error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "f1viewer",
		AllowedBackends: []keyring.BackendType{
			keyring.KWalletBackend,
			keyring.PassBackend,
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
		},
	})
	if err != nil {
		return err
	}
	session.ring = ring
	return nil
}

func (session *viewerSession) loadCredentials() error {
	if session.ring == nil {
		return errors.New("No keyring configured")
	}
	username, err := session.ring.Get("username")
	if err != nil {
		return fmt.Errorf("Could not get username: %w", err)
	}
	session.username = string(username.Data)

	password, err := session.ring.Get("password")
	if err != nil {
		return fmt.Errorf("Could not get password: %w", err)
	}
	session.password = string(password.Data)

	token, err := session.ring.Get("skylarkToken")
	if err != nil {
		session.logError("Could not get auth token: ", err)
	} else {
		session.authtoken = string(token.Data)
	}
	return nil
}

func (session *viewerSession) saveCredentials() error {
	if session.ring == nil {
		return errors.New("No keyring configured")
	}
	err := session.ring.Set(keyring.Item{
		Description: "F1TV username",
		Key:         "username",
		Data:        []byte(session.username),
	})
	if err != nil {
		return fmt.Errorf("[ERROR] could not save username %w", err)
	}

	err = session.ring.Set(keyring.Item{
		Description: "F1TV password",
		Key:         "password",
		Data:        []byte(session.password),
	})
	if err != nil {
		return fmt.Errorf("[ERROR] could not save password %w", err)
	}

	err = session.ring.Set(keyring.Item{
		Description: "F1TV auth token",
		Key:         "skylarkToken",
		Data:        []byte(session.authtoken),
	})
	if err != nil {
		return fmt.Errorf("[ERROR] could not save auth token %w", err)
	}
	return nil
}

func (session *viewerSession) removeCredentials() error {
	if session.ring == nil {
		return errors.New("No keyring configured")
	}
	err := session.ring.Remove("username")
	if err != nil {
		return fmt.Errorf("Could not remove username: %w", err)
	}
	session.username = ""

	err = session.ring.Remove("password")
	if err != nil {
		return fmt.Errorf("Could not remove password: %w", err)
	}
	session.password = ""

	err = session.ring.Remove("skylarkToken")
	if err != nil {
		return fmt.Errorf("Could not remove auth token: %w", err)
	}
	session.authtoken = ""

	return nil
}

func (session *viewerSession) updateUsername(username string) {
	session.username = username
}

func (session *viewerSession) updatePassword(password string) {
	session.password = password
}

func (session *viewerSession) updateToken(token string) {
	session.authtoken = token
}
