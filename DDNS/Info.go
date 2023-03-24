/*
 *     @Copyright
 *     @file: Info.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/25 上午1:46
 *     @last modified: 2023/3/25 上午1:46
 *
 *
 *
 */

// Package DDNS
// basic interfaces and tools for DDNS service
package DDNS

import (
	"GodDns/Util"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	FullName = "GodDns"
	Nickname = "go ddns"
)

const (
	owner  = "Equationzhao"
	repo   = "GodDns"
	Author = "Equationzhao"
)

var gits = []string{"github.com", "gitea.equationzhao.space:3000"}

// Version contains the major, minor, and patch versions.
type Version struct {
	major int
	minor int
	patch int
}

// RepoURLs return all repo urls
func RepoURLs() []string {
	var urls []string
	for _, git := range gits {
		urls = append(urls, fmt.Sprintf("git@%s:%s/%s.git", git, owner, repo))
	}
	return urls
}

// String returns a string representation of the version
// e.g. "1.2.3"
// The string is generated using the major, minor, and patch versions.
func (v Version) String() string {
	return strconv.Itoa(v.major) + "." + strconv.Itoa(v.minor) + "." + strconv.Itoa(v.patch)
}

// Info returns version info in format "v1.2.3"
func (v Version) Info() string {
	return fmt.Sprintf("v%s", v.String())
}

// Compare compares this version with another version.
// It returns 0 if they are equal, 1 if this version
// is greater than v2 and -1 if this version is less than v2.
func (v Version) Compare(v2 Version) int {
	if v.major > v2.major {
		return 1
	}
	if v.major < v2.major {
		return -1
	}
	if v.minor > v2.minor {
		return 1
	}
	if v.minor < v2.minor {
		return -1
	}
	if v.patch > v2.patch {
		return 1
	}
	if v.patch < v2.patch {
		return -1
	}
	return 0
}

// NowVersionInfo return version info
// like "GodDns (go ddns) version 0.1.0"
// FullName (Nickname) version major.minor.patch
func NowVersionInfo() string {
	return fmt.Sprintf("%s (%s) version %s", FullName, Nickname, NowVersion)
}

// NowVersion is current version of GodDns
var NowVersion = Version{
	major: 0,
	minor: 1,
	patch: 2,
}

// GetLatestVersionInfo get the latest version info from GitHub
// "https://api.github.com/repos/Equationzhao/GodDns/releases/latest
func GetLatestVersionInfo() (Version, string, error) {
	versionResponse := struct {
		Url       string `json:"url"`
		AssetsUrl string `json:"assets_url"`
		UploadUrl string `json:"upload_url"`
		HtmlUrl   string `json:"html_url"`
		Id        int    `json:"id"`
		Author    struct {
			Login             string `json:"login"`
			Id                int    `json:"id"`
			NodeId            string `json:"node_id"`
			AvatarUrl         string `json:"avatar_url"`
			GravatarId        string `json:"gravatar_id"`
			Url               string `json:"url"`
			HtmlUrl           string `json:"html_url"`
			FollowersUrl      string `json:"followers_url"`
			FollowingUrl      string `json:"following_url"`
			GistsUrl          string `json:"gists_url"`
			StarredUrl        string `json:"starred_url"`
			SubscriptionsUrl  string `json:"subscriptions_url"`
			OrganizationsUrl  string `json:"organizations_url"`
			ReposUrl          string `json:"repos_url"`
			EventsUrl         string `json:"events_url"`
			ReceivedEventsUrl string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"author"`
		NodeId          string    `json:"node_id"`
		TagName         string    `json:"tag_name"`
		TargetCommitish string    `json:"target_commitish"`
		Name            string    `json:"name"`
		Draft           bool      `json:"draft"`
		Prerelease      bool      `json:"prerelease"`
		CreatedAt       time.Time `json:"created_at"`
		PublishedAt     time.Time `json:"published_at"`
		Assets          []struct {
			Url      string      `json:"url"`
			Id       int         `json:"id"`
			NodeId   string      `json:"node_id"`
			Name     string      `json:"name"`
			Label    interface{} `json:"label"`
			Uploader struct {
				Login             string `json:"login"`
				Id                int    `json:"id"`
				NodeId            string `json:"node_id"`
				AvatarUrl         string `json:"avatar_url"`
				GravatarId        string `json:"gravatar_id"`
				Url               string `json:"url"`
				HtmlUrl           string `json:"html_url"`
				FollowersUrl      string `json:"followers_url"`
				FollowingUrl      string `json:"following_url"`
				GistsUrl          string `json:"gists_url"`
				StarredUrl        string `json:"starred_url"`
				SubscriptionsUrl  string `json:"subscriptions_url"`
				OrganizationsUrl  string `json:"organizations_url"`
				ReposUrl          string `json:"repos_url"`
				EventsUrl         string `json:"events_url"`
				ReceivedEventsUrl string `json:"received_events_url"`
				Type              string `json:"type"`
				SiteAdmin         bool   `json:"site_admin"`
			} `json:"uploader"`
			ContentType        string    `json:"content_type"`
			State              string    `json:"state"`
			Size               int       `json:"size"`
			DownloadCount      int       `json:"download_count"`
			CreatedAt          time.Time `json:"created_at"`
			UpdatedAt          time.Time `json:"updated_at"`
			BrowserDownloadUrl string    `json:"browser_download_url"`
		} `json:"assets"`
		TarballUrl string `json:"tarball_url"`
		ZipballUrl string `json:"zipball_url"`
		Body       string `json:"body"`
	}{}

	// https://api.github.com/repos/$owner$/$repo$/releases/latest
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	_, err := resty.New().R().SetResult(&versionResponse).Get(url)
	latest := Version{}
	if err != nil {
		return latest, "", err
	}

	versionStr := versionResponse.TagName
	_, err = fmt.Sscanf(versionStr, "v%d.%d.%d", &latest.major, &latest.minor, &latest.patch)

	if err != nil {
		return latest, "", err
	}

	os, arch := Util.OSDetect()
	for _, asset := range versionResponse.Assets {
		if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(os)) && strings.Contains(strings.ToLower(asset.Name), strings.ToLower(arch)) {
			// todo check compatibility, like x86 is compatible with amd64
			return latest, asset.BrowserDownloadUrl, nil
		}
	}

	return latest, "", NoCompatibleVersionError
}

// NoCompatibleVersion no compatible version of new release
type NoCompatibleVersion struct {
}

// NoCompatibleVersionError is the error of no compatible version
var NoCompatibleVersionError = NoCompatibleVersion{}

// Error return error info
// no compatible version
func (n NoCompatibleVersion) Error() string {
	return "no compatible version"
}

// CheckUpdate check if there is a new version
// if there is a new version, return true, the latest version, download url, nil
// if the new version is not compatible for current os, return true, the latest version, "", NoCompatibleVersionError
// if there is an error, return false, zero version, "", err
// if there is no new version, return false, zero version, "", nil
func CheckUpdate() (hasUpgrades bool, v Version, url string, err error) {
	latest, downloadURL, err := GetLatestVersionInfo()
	if err != nil {
		if errors.Is(err, NoCompatibleVersionError) {
			return true, latest, "", err
		} else {
			return false, v, "", err
		}
	}

	if NowVersion.Compare(latest) < 0 {
		// has upgrades
		return true, latest, downloadURL, nil
	} else {
		// no upgrades
		return false, v, "", nil
	}
}

// ? CheckPreRelease check if there is a new pre-release version

// FeedBackLink return feedback link
// https://github.com/$owner$/$repo$/issues
func FeedBackLink() string {
	return fmt.Sprintf("https://github.com/%s/%s/issues", owner, repo)
}

var feedbackEmail = "equationzhao@foxmail.com"

// FeedbackEmail return feedback email
func FeedbackEmail() string {
	return feedbackEmail
}

// Feedback return link and email
func Feedback() string {
	return fmt.Sprintf("Feedback: %s\nEmail: %s\n", FeedBackLink(), FeedbackEmail())
}
