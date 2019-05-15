package provider

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/mozvip/builds/version"
	"regexp"
	"strings"
)

func GitHubRelease(build Provider, currentVersion *version.Version) (remoteUrl string, buildVersion version.Version, err error){
	client := github.NewClient(nil)

	split := strings.Split(build.Name, "/")
	repositoryReleases, _, err := client.Repositories.ListReleases(context.Background(), split[0], split[1], nil)
	if err != nil {
		return remoteUrl, buildVersion, err
	}
	for _, release := range repositoryReleases {
		if release.PublishedAt.After(currentVersion.DateTime) {
			for _, asset := range release.Assets {
				re, err := regexp.Compile(build.AssetNameRegExp)
				if err == nil {
					if re.MatchString(*asset.Name) {
						return *asset.BrowserDownloadURL, version.NewDateTimeVersion(release.PublishedAt.Time), nil
					}
				}
			}
		}
	}

	return remoteUrl, buildVersion, err
}
