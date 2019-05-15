package main

import (
	"fmt"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/icons"
	"github.com/mozvip/builds/version"
	"gopkg.in/toast.v1"
	"log"
	"sync"
)

type CheckResult struct {
	Build builds.Build
	CurrentVersion version.Version
	AvailableVersion version.Version
	Err error
}

/*
func HandleBuildResult(results chan CheckResult, versions map[string]version.Version) {
	result := <- results
	if result.AvailableVersion.After(&result.CurrentVersion) {

		versions[result.Build.Name] = result.AvailableVersion

		notification := toast.Notification{
			AppID:   "Builds",
			Title:   fmt.Sprintf("%s : New build version %s was installed !", result.Build.Name, result.AvailableVersion),
			Message: fmt.Sprintf("A new build for %s was downloaded and installed...", result.Build.Name),
			ActivationArguments: result.Build.Location.Folder,
		}

		localIconPath := icons.GetIconForBuild(result.Build)
		if localIconPath != "" {
			notification.Icon = localIconPath
		}

		notification.Push()

		downloadCount ++
	}
}
*/
func checkBuilds(builds []builds.Build, versions map[string]version.Version) (downloadCount uint) {
	var wg sync.WaitGroup
	results := make(chan CheckResult, len(builds))
	for _, build := range builds {
		wg.Add(1)
		go checkBuild(results, versions, build, &wg)
		//go HandleBuildResult(results, versions)
	}
	wg.Wait()
	close(results)

	for result := range results {
		if result.AvailableVersion.After(&result.CurrentVersion) {

			versions[result.Build.Name] = result.AvailableVersion

			notification := toast.Notification{
				AppID:   "Builds",
				Title:   fmt.Sprintf("%s : New build version %s was installed !", result.Build.Name, result.AvailableVersion),
				Message: fmt.Sprintf("A new build for %s was downloaded and installed...", result.Build.Name),
				ActivationArguments: result.Build.Location.Folder,
			}

			localIconPath := icons.GetIconForBuild(result.Build)
			if localIconPath != "" {
				notification.Icon = localIconPath
			}

			notification.Push()

			downloadCount ++
		}
	}

	return downloadCount
}

func checkBuild(results chan CheckResult, versions map[string]version.Version, build builds.Build, wg *sync.WaitGroup) {
	defer wg.Done()

	currentVersion := versions[build.Name]
	newVersion, err := build.CheckBuild(&currentVersion)
	results <- CheckResult{Build: build, CurrentVersion:currentVersion, AvailableVersion:newVersion, Err: err}
}

func quit(versions map[string]version.Version) {
	err := version.SaveVersions(versions)
	if err != nil {
		log.Fatalln("Error saving versions", err)
	}
}

func findIconsForBuilds(builds []builds.Build) {
	var wg sync.WaitGroup
	for _, build := range builds {
		wg.Add(1)
		icons.CheckIcon(build)
		wg.Done()
	}
	wg.Wait()
}

func main() {

	localVersions, err := version.LoadVersions()
	if err != nil {
		log.Fatalln("Initialization failed", err)
	}
	defer quit(localVersions)

	from := builds.LoadBuildsFromFile("emulators.yaml")
	buildsFrom := builds.LoadBuildsFromFile("tools.yaml")

	from = append(from, buildsFrom...)

	findIconsForBuilds(from)

	downloadCount := checkBuilds(from, localVersions)

	if downloadCount > 0 {
		notification := toast.Notification{
			AppID: "Builds",
			Title: fmt.Sprintf("%d builds have been downloaded !", downloadCount),
			Message: fmt.Sprintf("%d builds have been downloaded and installed successfully", downloadCount),
		}
		notification.Push()
	}

}