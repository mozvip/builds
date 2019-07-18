package version

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type InstalledPackage struct {
	DisplayName string
	DisplayVersion string
}

type Version struct {
	FloatVersion float32	`yaml:"floatVersion,omitempty"`
	StringVersion string	`yaml:"stringVersion,omitempty"`
	DateTime time.Time		`yaml:"dateTime,omitempty"`
}

func (v Version) After(other *Version) bool {
	if v.StringVersion != "" {
		return v.StringVersion > other.StringVersion
	} else if v.FloatVersion != 0 {
		return v.FloatVersion > other.FloatVersion
	} else {
		return v.DateTime.After(other.DateTime)
	}
}

func (v Version) String() string {
	if v.StringVersion != "" {
		return v.StringVersion
	} else if v.FloatVersion != 0 {
		return fmt.Sprintf("%f", v.FloatVersion)
	} else {
		return v.DateTime.String()
	}

}

func NewStringVersion(version string) Version {
	var v Version
	v.StringVersion = version
	return v
}

func NewDateTimeVersion(dateTime time.Time) Version {
	var v Version
	v.DateTime = dateTime
	return v
}

func NewFloatVersion(version float32) Version {
	var v Version
	v.FloatVersion = version
	return v
}

func getVersionsFile() string {
	homeDir, _ := os.UserHomeDir()
	versionsFile := path.Join(homeDir, ".builds", "versions", "versions.yaml")
	return versionsFile
}

func LoadVersionsFromRegistry() ([]InstalledPackage, error) {

	var installedPackages []InstalledPackage

	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return installedPackages, err
	}
	defer k.Close()

	names, err := k.ReadSubKeyNames(-1)
	if err == nil {
		for _, packageName := range names {

			p := InstalledPackage{}

			packageKey, _ := registry.OpenKey(k, packageName, registry.QUERY_VALUE)
			val, _, err := packageKey.GetStringValue("DisplayName")
			if err == nil {
				p.DisplayName = val
			}
			vals, _, err := packageKey.GetStringValue("DisplayVersion")
			if err == nil {
				p.DisplayVersion = vals
			}

			installedPackages = append(installedPackages, p)
		}
	}

	return installedPackages, err
}

func LoadVersions() (map[string]Version, error) {

	installedPackages, err := LoadVersionsFromRegistry()
	if err != nil {
		log.Println(err)
	} else {
		log.Println(installedPackages)
	}

	versions := make(map[string]Version)

	versionsFile := getVersionsFile()

	_, e := os.Stat(versionsFile)
	if e != nil && os.IsNotExist(e) {
		return versions, nil
	}
	file, err := os.Open(versionsFile)
	if err != nil {
		return versions, err
	}
	defer file.Close()

	info, err := file.Stat()
	fileData := make([]byte, info.Size())
	_, err = file.Read(fileData)

	err = yaml.Unmarshal([]byte(fileData), &versions)
	return versions, nil
}

func SaveVersions(versions map[string]Version) error {
	out, err := yaml.Marshal(versions)
	if err == nil {
		err = ioutil.WriteFile(getVersionsFile(), out, 0644)
	}
	return err
}