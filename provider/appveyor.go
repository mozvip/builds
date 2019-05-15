package provider

import (
	"encoding/json"
	"fmt"
	"github.com/mozvip/builds/version"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

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

func AppVeyor(build Provider, currentVersion *version.Version) (localFile string, buildVersion version.Version, err error) {
	apiUrl := fmt.Sprintf("https://ci.appveyor.com/api/projects/%s/branch/%s", build.Name, build.Branch)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return "", version.NewStringVersion(""), err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		var result AppVeyorResult
		json.Unmarshal(bytes, &result)
		for _,job := range result.Build.Jobs {

			if build.JobRegExp != "" {
				matched, _ := regexp.MatchString(build.JobRegExp, job.Name)
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
						if artifact.Name == build.DeploymentName || artifact.FileName == build.DeploymentName {
							return fmt.Sprintf("https://ci.appveyor.com/api/buildjobs/%s/artifacts/%s", job.JobId, artifact.FileName), version.NewDateTimeVersion(job.Finished), nil
						}
					}
				}
			}
		}
	}

	return "", version.NewStringVersion(""), err

}