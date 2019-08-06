package provider

import (
	"fmt"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/search"
	"github.com/mozvip/builds/version"
	"log"
	"os/exec"
	"regexp"
)

type ChocolateyProvider struct{}

func (ChocolateyProvider) Init() {
	log.Println("ChocolateyProvider init : TODO")
}

func (ChocolateyProvider) Update() {
	log.Println("ChocolateyProvider update : TODO")
}

func (ChocolateyProvider) CanHandle(buildType string) bool {
	return buildType == "chocolatey"
}

func (ChocolateyProvider) Search(packageName string) []search.Result {
	return []search.Result{}
}

func (ChocolateyProvider) NeedsInstallLocation() bool {
	return false
}

func (ChocolateyProvider) DownloadBuild(providerData *builds.ProviderData, currentVersion *version.Version) search.Result {

	var installedVersion, availableVersion version.Version

	// get installed version
	commandOutput, err := exec.Command("chocolatey", "list", "--local-only", providerData.Name).CombinedOutput()
	if err == nil {
		installedVersion = extractVersion(providerData.Name, commandOutput)
	} else {
		log.Println("Error invoking chocolatey", err)
		log.Println(string(commandOutput))
		return search.Error(err)
	}

	// get available version
	commandOutput, err = exec.Command("chocolatey", "list", providerData.Name).CombinedOutput()
	if err == nil {
		availableVersion = extractVersion(providerData.Name, commandOutput)
	} else {
		log.Println("Error invoking chocolatey", err)
		log.Println(string(commandOutput))
		return search.Error(err)
	}

	if availableVersion.After(&installedVersion) {
		output, err := exec.Command("chocolatey", "upgrade", providerData.Name, "-y").CombinedOutput()
		if err != nil {
			return search.Error(err)
		}
		fmt.Println(string(output))
	}

	return search.Installed(availableVersion)
}

func extractVersion(packageName string, commandOutput []byte) version.Version {
	re := regexp.MustCompile(fmt.Sprintf(".*%s\\s+([\\w\\.]+).*", packageName))
	submatch := re.FindStringSubmatch(string(commandOutput))
	var v version.Version
	if len(submatch) > 0 {
		v = version.NewStringVersion(submatch[1])
	} else {
		log.Printf("Could not determine version of %s\n", packageName)
	}
	return v
}
