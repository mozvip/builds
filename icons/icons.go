package icons

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/tools/files"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type BuildIcon struct {
	Name string
	LocalIconPath string
}

type IconManager struct {
	IconFolder string
}

func New(configDir string) IconManager {
	return IconManager{IconFolder: path.Join(configDir, "icons")}
}

func (i IconManager) getIconFileName(build builds.Build) string {
	iconsFile := path.Join(i.IconFolder, build.Name + ".png")
	return iconsFile
}

func (i IconManager)  GetIconForBuild(build builds.Build) string {
	iconsFile := i.getIconFileName(build)
	_, err := os.Stat(iconsFile)
	if err != nil && os.IsNotExist(err) {
		return ""
	}
	return iconsFile
}

func (i IconManager)  downloadIcon(imageUrl string, baseURL *url.URL, iconFile string) (err error) {
	if !strings.HasPrefix(imageUrl, "http") {
		imageUrl, err = files.MakeAbsoluteUrl(imageUrl, baseURL)
		if err != nil {
			return err
		}
	}
	localFile, downloadErr := files.DownloadFile(imageUrl, i.IconFolder)
	if downloadErr == nil && localFile != iconFile {
		os.Rename(localFile, iconFile)
	}
	return downloadErr
}

func (i IconManager) CheckIcon(build builds.Build) error {

	iconFile := i.getIconFileName(build)
	_, err := os.Stat(iconFile)
	if err == nil {
		// icon is there
		return nil
	}

	if build.Provider.Type == "chocolatey" {
		resp, err := http.Get("https://chocolatey.org/packages/" + build.Provider.Name)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		document, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return err
		}
		logo := document.Find("img.logo").First()
		imageUrl, exists := logo.Attr("src")
		if !exists {
			return nil
		}
		if strings.HasSuffix(imageUrl, "/Content/Images/packageDefaultIcon.png") {
			return nil
		}
		if strings.HasSuffix(imageUrl, ".svg") {
			// TODO: TEST
			return nil
		}
		i.downloadIcon(imageUrl, resp.Request.URL, iconFile)
	} else if build.Provider.Type == "httpLink" {
		requestUrl, err := url.Parse(build.Provider.Url)
		if err != nil {
			return err
		}
		// request root url
		resp, httpErr := http.Get(requestUrl.Scheme + "://" + requestUrl.Host)
		if httpErr != nil {
			return httpErr
		}
		defer resp.Body.Close()

		document, goQueryErr := goquery.NewDocumentFromReader(resp.Body)
		if goQueryErr != nil {
			return goQueryErr
		}

		selection := document.Find("link[rel=icon]")
		if selection.Size() > 0 {
			val, exists := selection.First().Attr("href")
			if exists {
				i.downloadIcon(val, requestUrl, iconFile)
			}
		} else {
			// TODO
		}

	} else {
		// TODO: try to find a matching icon online
	}

	return nil
}