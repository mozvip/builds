package builds

import (
	"github.com/mozvip/builds/provider"
	"github.com/mozvip/builds/tools/files"
	"github.com/mozvip/builds/tools/sevenzip"
	"github.com/mozvip/builds/version"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Build struct {
	Name           string
	Provider       provider.Provider
	Location       Location
	PostInstallCmd string `yaml:"postInstallCmd"`
}

type Location struct {
	Type                 string
	Folder               string
	SuppressParentFolder bool `yaml:"suppressParentFolder"`
	AddToPath            bool `yaml:"addToPath"`
}

func (build Build) executePostBuildCommand(localFile string) (err error) {
	split := strings.Split(build.PostInstallCmd, " ")
	var commandLine []string
	for _, string := range split {
		string = strings.Replace(string, "${ARTIFACT}", localFile, -1)
		if build.Location.Folder != "" {
			string = strings.Replace(string, "${LOCATION}", localFile, -1)
		}
		commandLine = append(commandLine, string)
	}
	var combinedOutputBytes []byte
	combinedOutputBytes, err = exec.Command(commandLine[0], commandLine[1:]...).CombinedOutput()
	if err != nil {
		log.Println("Error occurred : ", strings.TrimSpace(string(combinedOutputBytes)), err)
	}
	return err
}

func (build Build) CheckBuild(currentVersion *version.Version) (version.Version, error) {

	if build.Provider.Type != "chocolatey" && build.Location.Folder == "" && build.PostInstallCmd == "" {
		log.Printf("%s has not location set, skipping installation", build.Name)
		return *currentVersion, nil
	}

	log.Printf("Checking for new build of %s, current version = %s", build.Name, currentVersion)

	var remoteUrl string
	var err error
	var newVersion version.Version

	switch build.Provider.Type {
	case "httpLink":
		remoteUrl, newVersion, err = provider.HttpLink(build.Provider, currentVersion)
	case "appVeyor":
		remoteUrl, newVersion, err = provider.AppVeyor(build.Provider, currentVersion)
	case "githubRelease":
		remoteUrl, newVersion, err = provider.GitHubRelease(build.Provider, currentVersion)
	case "chocolatey":
		newVersion, err = provider.Chocolatey(build.Provider, currentVersion)
	}

	if err != nil {
		log.Printf("Error checking build for %s : %s\n", build.Name, err)
		return *currentVersion, err
	}

	if remoteUrl == "" {
		return *currentVersion, err
	}

	log.Printf("Downloading %s for new build of %s at version %s\n", remoteUrl, build.Name, newVersion)
	localFile, err := files.DownloadFile(remoteUrl, os.TempDir())
	if err != nil {
		return *currentVersion, err
	}
	log.Printf("New build for %s was downloaded, installing from local file %s\n", build.Name, localFile)
	if strings.HasSuffix(localFile, ".7z") || strings.HasSuffix(localFile, ".zip") {
		entries, err := sevenzip.ReadEntries(localFile)
		var commonFolderName string
		if err != nil {
			return *currentVersion, err
		}
		// is first entry a folder ?
		if strings.HasPrefix(entries[0].Attr, "D") {
			commonFolderName = entries[0].Name
			// check if all remaining entries are in this folder
			for _, entry := range entries[1:] {
				if !strings.HasPrefix(entry.Name, commonFolderName) {
					commonFolderName = ""
					break
				}
			}
		}

		// identify windows executables from archive
		/*
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name, ".exe") {

				}
			}
		*/

		if commonFolderName != "" && build.Location.SuppressParentFolder {
			unzipErr := sevenzip.Uncompress(localFile, os.TempDir())
			if unzipErr == nil {
				sourceFolder := path.Join(os.TempDir(), commonFolderName)
				err = files.MoveFolder(sourceFolder, build.Location.Folder)
			}
		} else {
			err = sevenzip.Uncompress(localFile, build.Location.Folder)
		}
		if err == nil {
			os.Remove(localFile)
		}
	}

	if build.PostInstallCmd != "" {
		err = build.executePostBuildCommand(localFile)
		if err != nil {
			return *currentVersion, err
		} else {
			os.Remove(localFile)
		}
	}

	if build.Location.AddToPath {
		path := os.Getenv("PATH")
		paths := strings.Split(path, ";")
		alreadyInPath := false
		for _, value := range paths {
			if value == build.Location.Folder {
				alreadyInPath = true
				break
			}
		}
		if !alreadyInPath {
			pathValue := path + ";" + build.Location.Folder
			combinedOutputBytes, err := exec.Command("SETX", "PATH", pathValue).CombinedOutput()
			if err != nil {
				log.Println(string(combinedOutputBytes))
			}
		}
	}

	return newVersion, nil
}

func LoadBuildsFromFile(fileName string) []Build {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	info, err := file.Stat()
	fileData := make([]byte, info.Size())
	_, err = file.Read(fileData)
	file.Close()

	var builds []Build
	err = yaml.Unmarshal([]byte(fileData), &builds)

	return builds
}
