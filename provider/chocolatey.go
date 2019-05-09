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
		re := regexp.MustCompile(fmt.Sprintf(".*%s\\s+([\\w\\.]+).*", build.Name))
		submatch := re.FindStringSubmatch(string(commandOutput))
		installedVersion = version.NewStringVersion(submatch[1])
	} else {
		log.Println("Error invoking chocolatey", err)
		log.Println(string(commandOutput))
	}

	// get available version
	commandOutput, err = exec.Command("chocolatey", "list", build.Name).CombinedOutput()
	if err == nil {
		re := regexp.MustCompile(fmt.Sprintf(".*%s\\s+([\\w\\.]+).*", build.Name))
		submatch := re.FindStringSubmatch(string(commandOutput))
		availableVersion = version.NewStringVersion(submatch[1])
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
