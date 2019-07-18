package provider

import (
	"encoding/json"
	"fmt"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/search"
	"github.com/mozvip/builds/version"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type AppVeyorProvider struct {
}

type AppVeyorResult struct {
	Build struct {
		Jobs []struct {
			JobId string
			Name string
			Status string
			Finished time.Time
		}
	}
}

type AppVeyorArtifact struct {
	FileName string
	Name string
	Type string
	Size uint32
}

func (AppVeyorProvider) Init() {
}

func (AppVeyorProvider) Update() {
}

func (AppVeyorProvider) CanHandle(buildType string) bool {
	return buildType == "appVeyor"
}

func (AppVeyorProvider) Search(packageName string) []search.SearchResult {
	return []search.SearchResult{}
}

func (AppVeyorProvider) NeedsInstallLocation() bool {
	return true
}

func (AppVeyorProvider) DownloadBuild(build *builds.Build, currentVersion *version.Version) (version.Version, error) {
	apiUrl := fmt.Sprintf("https://ci.appveyor.com/api/projects/%s/branch/%s", build.Provider.Name, build.Provider.Branch)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return version.NewStringVersion(""), err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		var result AppVeyorResult
		json.Unmarshal(bytes, &result)
		for _,job := range result.Build.Jobs {

			if build.Provider.JobRegExp != "" {
				matched, _ := regexp.MatchString(build.Provider.JobRegExp, job.Name)
				if !matched {
					// skip this job
					break
				}
			}

			if job.Status == "success" && job.Finished.After(currentVersion.DateTime){
				// get build artifacts
				jobArtifactsUrl := fmt.Sprintf("https://ci.appveyor.com/api/buildjobs/%s/artifacts", job.JobId)
				artifactResponse, errArtifacts := http.Get(jobArtifactsUrl)
				if errArtifacts == nil {
					defer artifactResponse.Body.Close()
					bytes, _ := ioutil.ReadAll(artifactResponse.Body)
					var artifacts []AppVeyorArtifact
					json.Unmarshal(bytes, &artifacts)
					for _, artifact := range artifacts {
						if artifact.Name == build.Provider.DeploymentName || artifact.FileName == build.Provider.DeploymentName {
							url := fmt.Sprintf("https://ci.appveyor.com/api/buildjobs/%s/artifacts/%s", job.JobId, artifact.FileName)
							err = build.DownloadBuildFromURL(url)
							return version.NewDateTimeVersion(job.Finished), err
						}
					}
				}
			}
		}
	}

	return version.NewStringVersion(""), err

}
