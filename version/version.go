package version

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
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

func LoadVersions() map[string]Version {
	versions := make(map[string]Version)

	_, e := os.Stat("updates.yaml")
	if e != nil && os.IsNotExist(e) {
		return versions
	}
	file, err := os.Open("updates.yaml")
	if err != nil {
		log.Println("Could not read updates.yaml", err)
		return versions
	}
	defer file.Close()

	info, err := file.Stat()
	fileData := make([]byte, info.Size())
	_, err = file.Read(fileData)

	err = yaml.Unmarshal([]byte(fileData), &versions)

	return versions
}

func SaveVersions(versions map[string]Version) error {
	out, err := yaml.Marshal(versions)
	if err == nil {
		err = ioutil.WriteFile("updates.yaml", out, 0644)
	}
	return err
}