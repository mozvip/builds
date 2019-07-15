package provider

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/version"
	"regexp"
	"strings"
)

type GitHubProvider struct {}

func (GitHubProvider) Init() {
}

func (GitHubProvider) CanHandle(buildType string) bool {
	return buildType == "githubRelease"
}

func (GitHubProvider) NeedsInstallLocation() bool {
	return true
}

func (GitHubProvider) DownloadBuild(build *builds.Build, currentVersion *version.Version) (version.Version, error) {
	client := github.NewClient(nil)

	split := strings.Split(build.Provider.Name, "/")
	repositoryReleases, _, err := client.Repositories.ListReleases(context.Background(), split[0], split[1], nil)
	if err == nil {
		for _, release := range repositoryReleases {
			if release.PublishedAt.After(currentVersion.DateTime) {
				for _, asset := range release.Assets {
					re, err := regexp.Compile(build.Provider.AssetNameRegExp)
					if err == nil {
						if re.MatchString(*asset.Name) {
							build.DownloadBuildFromURL(*asset.BrowserDownloadURL)
							return version.NewDateTimeVersion(release.PublishedAt.Time), nil
						}
					}
				}
			}
		}
	}
	return *currentVersion, err
}
