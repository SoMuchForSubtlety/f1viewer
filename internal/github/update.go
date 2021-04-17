package github

import (
	"encoding/json"
	"net/http"
)

const githubURL = "https://api.github.com/repos/SoMuchForSubtlety/F1viewer/releases/latest"

type Release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Body       string `json:"body"`
}

func CheckUpdate(version string) (Release, bool, error) {
	resp, err := http.Get(githubURL)
	if err != nil {
		return Release{}, false, err
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return Release{}, false, err
	}

	new := release.TagName != version &&
		release.TagName != "v"+version &&
		version != "dev"

	return release, new, nil
}
