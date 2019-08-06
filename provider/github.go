package provider

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/search"
	"github.com/mozvip/builds/version"
	"regexp"
	"strings"
)

type GitHubProvider struct {}

func (GitHubProvider) Init() {
}

func (GitHubProvider) Update() {
}


func (GitHubProvider) CanHandle(buildType string) bool {
	return buildType == "githubRelease"
}

func (GitHubProvider) NeedsInstallLocation() bool {
	return true
}

func (GitHubProvider) Search(packageName string) []search.Result {
	return []search.Result{}
}

func (GitHubProvider) DownloadBuild(providerData *builds.ProviderData, currentVersion *version.Version) search.Result {
	client := github.NewClient(nil)

	split := strings.Split(providerData.Name, "/")
	repositoryReleases, _, err := client.Repositories.ListReleases(context.Background(), split[0], split[1], nil)
	if err != nil {
		return search.Error(err)
	}
	for _, release := range repositoryReleases {
		if release.PublishedAt.After(currentVersion.DateTime) {
			for _, asset := range release.Assets {
				re, err := regexp.Compile(providerData.AssetNameRegExp)
				if err == nil {
					if re.MatchString(*asset.Name) {
						return search.New("", version.NewDateTimeVersion(release.PublishedAt.Time), *asset.BrowserDownloadURL)
					}
				}
			}
		}
	}
	return search.None()
}
