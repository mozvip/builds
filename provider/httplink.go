package provider

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/tools/files"
	"github.com/mozvip/builds/version"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type HttpLinkProvider struct {
}

func (HttpLinkProvider) Init() {
}

func (HttpLinkProvider) CanHandle(buildType string) bool {
	return buildType == "httpLink"
}

func (HttpLinkProvider) NeedsInstallLocation() bool {
	return true
}

func (HttpLinkProvider) DownloadBuild(build *builds.Build, currentVersion *version.Version) (version.Version, error) {

	link := build.Provider.Url

	var availableVersion version.Version
	if build.Provider.VersionSelector != "" || build.Provider.LinkSelector != "" {
		link, availableVersion, _ = determineLinkAndVersion(build.Provider.Url, build.Provider.LinkSelector, build.Provider.VersionSelector, build.Provider.VersionRegExp)
	} else {
		resp, err := http.Head(build.Provider.Url)
		if err == nil {
			buildTime, _ := http.ParseTime(resp.Header.Get("Last-Modified"))
			availableVersion = version.NewDateTimeVersion(buildTime)
		}
	}

	if !availableVersion.After(currentVersion) {
		return availableVersion, nil
	}

	log.Printf("Downloading %s for new build of %s at version %s\n", link, build.Name, availableVersion)
	e := build.DownloadBuildFromURL(link)

	return availableVersion, e
}

func determineLinkAndVersion(downloadUrl string, linkSelector string, versionSelector string, versionRegExp string) (link string, linkVersion version.Version, err error) {

	res, err := http.Get(downloadUrl)
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}

		dom, errParse := goquery.NewDocumentFromReader(res.Body)
		if errParse == nil {

			selection := dom.Find(linkSelector)
			var exists bool
			link, exists = selection.Attr("href")
			if exists {
				if !strings.HasPrefix(link, "http") {
					link, err = files.MakeAbsoluteUrl(link, res.Request.URL)
					if err != nil {
						return "", version.Version{}, nil
					}
				}

				if versionSelector != "" {

					versionString := dom.Find(versionSelector).First().Text()

					if versionRegExp != "" {
						regex, errRegexp := regexp.Compile(versionRegExp)
						if errRegexp == nil {
							matches := regex.FindStringSubmatch(versionString)
							versionString = matches[1]
						}
					}

					linkVersion = version.NewStringVersion(versionString)
				}
			}
		}
	}

	return link, linkVersion, err
}