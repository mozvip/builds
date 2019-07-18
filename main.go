package main

import (
	"flag"
	"fmt"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/icons"
	"github.com/mozvip/builds/provider"
	"github.com/mozvip/builds/search"
	"github.com/mozvip/builds/version"
	"gopkg.in/toast.v1"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

type CheckResult struct {
	Build builds.Build
	CurrentVersion version.Version
	AvailableVersion version.Version
	Err error
}

var buildProviders []provider.BuildProvider

func checkBuilds(builds []builds.Build, versions map[string]version.Version) (downloadCount uint) {

	var wg sync.WaitGroup
	results := make(chan CheckResult, len(builds))
	for _, build := range builds {
		wg.Add(1)
		go checkBuild(results, buildProviders, versions, build, &wg)
	}
	wg.Wait()
	close(results)

	for result := range results {

		if result.Err != nil {
			log.Printf("Error checking build for %s : %s\n", result.Build.Name, result.Err)
		} else {
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

	}

	return downloadCount
}

func checkBuild(results chan CheckResult, buildProviders []provider.BuildProvider, versions map[string]version.Version, build builds.Build, wg *sync.WaitGroup) {
	defer wg.Done()

	currentVersion := versions[build.Name]
	newVersion, err := provider.CheckBuild(&build, buildProviders, &currentVersion)
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

	homeDir, err := os.UserHomeDir()
	buildsDir := path.Join(homeDir, ".builds")

	var action = flag.String("mode", "update", "action to execute")
	var pack = flag.String("package", "all", "package to execute action on")

	flag.Parse()

	buildProviders = provider.Init()

	infos, err := ioutil.ReadDir(buildsDir)

	var buildsToCheck []builds.Build
	for _, value := range infos {
		if *pack != "all" && len(buildsToCheck) > 0 {
			break
		}
		if strings.HasSuffix(value.Name(), ".yaml") {
			buildsFromFile := builds.LoadBuildsFromFile(path.Join(buildsDir, value.Name()))
			if *pack == "all" {
				buildsToCheck = append(buildsToCheck, buildsFromFile...)
			} else {
				for _, value := range buildsFromFile {
					if value.Name == *pack {
						buildsToCheck = append(buildsToCheck, value)
						break
					}
				}
			}
		}
	}

	if *action == "update" {
		findIconsForBuilds(buildsToCheck)
		checkBuilds(buildsToCheck, localVersions)
	} else if *action == "search" {

		var results []search.SearchResult

		for _, provider := range buildProviders {
			results = append(results, provider.Search(*pack)...)
		}

		log.Println(results)
		
	}
}