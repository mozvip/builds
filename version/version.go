package version

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type Version struct {
	Version string
	DateTime time.Time	`yaml:"dateTime"`
}

func (v Version) After(otherVersion *Version) bool {

	if otherVersion == nil {
		return true
	}

	if v.Version != "" {
		return v.Version > otherVersion.Version
	} else {
		return v.DateTime.After(otherVersion.DateTime)
	}
}

func (v Version) String() string {
	if v.Version != "" {
		return v.Version
	} else {
		return v.DateTime.String()
	}
}

func NewStringVersion(version string) Version {
	var v Version
	v.Version = version
	return v
}

func NewDateTimeVersion(dateTime time.Time) Version {
	var v Version
	v.DateTime = dateTime
	return v
}

func getVersionsFile() string {
	homeDir, _ := os.UserHomeDir()
	versionsFile := path.Join(homeDir, ".builds", "versions.yaml")
	return versionsFile
}

func LoadVersions() (map[string]Version, error) {
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