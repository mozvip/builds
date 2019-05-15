package provider

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/mozvip/builds/tools/files"
	"github.com/mozvip/builds/version"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func HttpLink(provider Provider, currentVersion *version.Version) (remoteUrl string, buildVersion version.Version, err error) {

	link := provider.Url

	if provider.VersionSelector != "" || provider.LinkSelector != "" {
		link, buildVersion, _ = determineLinkAndVersion(provider.Url, provider.LinkSelector, provider.VersionSelector, provider.VersionRegExp)
	} else {
		resp, err := http.Head(provider.Url)
		if err == nil {
			buildTime, _ := http.ParseTime(resp.Header.Get("Last-Modified"))
			buildVersion = version.NewDateTimeVersion(buildTime)
		}
	}

	if !buildVersion.After(currentVersion) {
		return "", buildVersion, nil
	}

	return link, buildVersion, err
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