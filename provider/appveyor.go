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

func (AppVeyorProvider) Search(packageName string) []search.Result {
	return []search.Result{}
}

func (AppVeyorProvider) NeedsInstallLocation() bool {
	return true
}

func (AppVeyorProvider) DownloadBuild(providerData *builds.ProviderData, currentVersion *version.Version) search.Result {
	apiUrl := fmt.Sprintf("https://ci.appveyor.com/api/projects/%s/branch/%s", providerData.Name, providerData.Branch)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return search.Error(err)
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		var result AppVeyorResult
		json.Unmarshal(bytes, &result)
		for _,job := range result.Build.Jobs {

			if providerData.JobRegExp != "" {
				matched, _ := regexp.MatchString(providerData.JobRegExp, job.Name)
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
					return func() search.Result {
						defer artifactResponse.Body.Close()
						bytes, _ := ioutil.ReadAll(artifactResponse.Body)
						var artifacts []AppVeyorArtifact
						json.Unmarshal(bytes, &artifacts)
						for _, artifact := range artifacts {
							if artifact.Name == providerData.DeploymentName || artifact.FileName == providerData.DeploymentName {
								url := fmt.Sprintf("https://ci.appveyor.com/api/buildjobs/%s/artifacts/%s", job.JobId, artifact.FileName)
								return search.New(artifact.Name, version.NewDateTimeVersion(job.Finished), url)
							}
						}
						return search.None()
					}()
				}
			}
		}
	}

	return search.None()

}
