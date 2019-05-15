package provider

import (
	"fmt"
	"github.com/mozvip/builds/version"
	"log"
	"os/exec"
	"regexp"
)

func Chocolatey(build Provider, currentVersion *version.Version) (buildVersion version.Version, err error) {

	var installedVersion, availableVersion version.Version

	// get installed version
	commandOutput, err := exec.Command("chocolatey", "list", "--local-only", build.Name).CombinedOutput()
	if err == nil {
		installedVersion = extractVersion(build, commandOutput)
	} else {
		log.Println("Error invoking chocolatey", err)
		log.Println(string(commandOutput))
	}

	// get available version
	commandOutput, err = exec.Command("chocolatey", "list", build.Name).CombinedOutput()
	if err == nil {
		availableVersion = extractVersion(build, commandOutput)
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

func extractVersion(build Provider, commandOutput []byte) version.Version {
	re := regexp.MustCompile(fmt.Sprintf(".*%s\\s+([\\w\\.]+).*", build.Name))
	submatch := re.FindStringSubmatch(string(commandOutput))
	v := version.NewStringVersion(submatch[1])
	return v
}
