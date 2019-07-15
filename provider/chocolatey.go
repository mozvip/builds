package provider

import (
	"fmt"
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/version"
	"log"
	"os/exec"
	"regexp"
)

type ChocolateyProvider struct{}

func (ChocolateyProvider) Init() {
	log.Println("ChocolateyProvider init : TODO")
}

func (ChocolateyProvider) CanHandle(buildType string) bool {
	return buildType == "chocolatey"
}

func (ChocolateyProvider) NeedsInstallLocation() bool {
	return false
}

func (ChocolateyProvider) DownloadBuild(build *builds.Build, currentVersion *version.Version) (version.Version, error) {

	var installedVersion, availableVersion version.Version

	// get installed version
	commandOutput, err := exec.Command("chocolatey", "list", "--local-only", build.Name).CombinedOutput()
	if err == nil {
		installedVersion = extractVersion(build.Provider.Name, commandOutput)
	} else {
		log.Println("Error invoking chocolatey", err)
		log.Println(string(commandOutput))
	}

	// get available version
	commandOutput, err = exec.Command("chocolatey", "list", build.Name).CombinedOutput()
	if err == nil {
		availableVersion = extractVersion(build.Provider.Name, commandOutput)
	} else {
		log.Println("Error invoking chocolatey", err)
		log.Println(string(commandOutput))
	}

	if availableVersion.After(&installedVersion) {
		output, err := exec.Command("chocolatey", "upgrade", build.Name, "-y").CombinedOutput()
		if err == nil {

		}
		fmt.Println(string(output))
	}

	return availableVersion, err
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
