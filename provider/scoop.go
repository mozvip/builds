package provider

import (
	"bytes"
	"fmt"
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

func (s ScoopProvider) Search(packageName string) []search.SearchResult {

	var results []search.SearchResult
	output, err := commands.RunCommand(exec.Command("scoop", "search", packageName))
	if err == nil {
		re := regexp.MustCompile("\n\\s+(\\S+)\\s+\\(([^\\)]+)\\).*")
		searchResults := re.FindAllStringSubmatch(string(output), -1)
		for _, value := range searchResults {
			if value[1] == packageName {
				version := version.NewStringVersion(value[2])
				results = append(results, search.New(value[1], version))
			}
		}
	}

	return results
}

func (s ScoopProvider) NeedsInstallLocation() bool {
	return false
}

func (s ScoopProvider) DownloadBuild(build *builds.Build, currentVersion *version.Version) (version.Version, error) {

	var installedVersion, availableVersion version.Version

	installedVersion = s.InstalledPackages[build.Provider.Name]

	commandOutput, err := exec.Command("scoop", "search", build.Provider.Name).CombinedOutput()
	if err == nil {
		availableVersion = scoopExtractVersion(build.Provider.Name, commandOutput)
		if availableVersion.After(&installedVersion) {
			commandOutput, err = exec.Command("scoop", "upgrade", build.Provider.Name).CombinedOutput()
			if err == nil {
				return availableVersion, err
			}
		}
	}

	return installedVersion, err
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

func scoopExtractVersion(packageName string, commandOutput []byte) version.Version {
	re := regexp.MustCompile(fmt.Sprintf("\\s+%s\\s+\\(([\\w\\.]+)\\).*", packageName))
	submatch := re.FindStringSubmatch(string(commandOutput))
	if len(submatch) > 0 {
		return version.NewStringVersion(submatch[1])
	} else {
		return version.NewStringVersion("")
	}
}