package provider

import (
	"bytes"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/search"
	"github.com/mozvip/builds/tools/commands"
	"github.com/mozvip/builds/version"
	"log"
	"os/exec"
	"regexp"
)

type ScoopProvider struct {
	InstalledPackages map[string]version.Version
}

func (s ScoopProvider) CanHandle(buildType string) bool {
	return buildType == "scoop"
}

func (s ScoopProvider) Search(packageName string) []search.Result {

	var results []search.Result
	output, err := commands.RunCommand(exec.Command("scoop", "search", packageName))
	if err == nil {
		re := regexp.MustCompile("\n\\s+(\\S+)\\s+\\(([^\\)]+)\\).*")
		searchResults := re.FindAllStringSubmatch(string(output), -1)
		for _, value := range searchResults {
			if value[1] == packageName {
				version := version.NewStringVersion(value[2])
				results = append(results, search.New(value[1], version, ""))
			}
		}
	}

	return results
}

func (s ScoopProvider) NeedsInstallLocation() bool {
	return false
}

func (s ScoopProvider) InstallPackage(packageName string) search.Result {

	installedVersion, installed := s.InstalledPackages[packageName]

	results := s.Search(packageName)

	if len(results) == 1 {
		availableVersion := results[0].Version
		if availableVersion.After(&installedVersion) {
			var err error
			if installed {
				_, err = exec.Command("scoop", "update", packageName).CombinedOutput()
			} else {
				_, err = exec.Command("scoop", "install", packageName).CombinedOutput()
			}
			if err != nil {
				return search.Error(err)
			}
		}
		return search.Installed(availableVersion)
	}

	return search.None()
}

func (s ScoopProvider) DownloadBuild(providerData *builds.ProviderData, currentVersion *version.Version) search.Result {

	return s.InstallPackage(providerData.Name)

}

func (s *ScoopProvider) Update() {
	output, _ := exec.Command("scoop", "update").CombinedOutput()
	log.Println(string(output))
}

func (s *ScoopProvider) Init() {
	path, e := exec.LookPath("scoop")
	if e == nil {
		log.Printf("scoop located at %s\n", path)
	} else {
		// scoop was not found : install it
		output, _ := exec.Command("powershell", "iex", "(new-object net.webclient).downloadstring('https://get.scoop.sh')").CombinedOutput()
		log.Println(string(output))
	}

	// scoop bucket add main

	command := exec.Command("scoop", "list")

	var b bytes.Buffer
	command.Stdout = &b
	err := command.Run()

	if err == nil {
		s.InstalledPackages = make(map[string]version.Version)
		re := regexp.MustCompile("\n\\s+(\\S+)\\s+([\\w\\.]+).*")
		allApps := re.FindAllStringSubmatch(string(b.Bytes()), -1)
		for _, value := range allApps {
			s.InstalledPackages[value[1]] = version.NewStringVersion(value[2])
		}
	}
}