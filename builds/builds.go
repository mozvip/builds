package builds

import (
	"github.com/mozvip/builds/tools/files"
	"github.com/mozvip/builds/tools/sevenzip"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type ProviderData struct {
	Type string
	Url string
	LinkSelector string		`yaml:"linkSelector"`
	VersionSelector string  `yaml:"versionSelector"`
	VersionRegExp string   	`yaml:"versionRegExp"`
	Name string
	DeploymentName string	`yaml:"deploymentName"`
	Branch string
	// for githubRelease
	AssetNameRegExp	string	`yaml:"assetNameRegExp"`
	// for appVeyor
	JobRegExp string		`yaml:"jobRegExp"`
}


type Build struct {
	Name           string
	Provider       ProviderData
	Location       Location
	PostInstallCmd string `yaml:"postInstallCmd"`
}

type Location struct {
	Type                 string
	Folder               string
	SuppressParentFolder bool `yaml:"suppressParentFolder"`
	AddToPath            bool `yaml:"addToPath"`
}

func (build Build) DownloadBuildFromURL(remoteURL string) error {
	downloadedFile, err := files.DownloadFile(remoteURL, os.TempDir())
	if err != nil {
		return err
	}
	log.Printf("New build for %s was downloaded, installing from local file %s\n", build.Name, downloadedFile)
	if strings.HasSuffix(downloadedFile, ".7z") || strings.HasSuffix(downloadedFile, ".zip") {
		var entries []sevenzip.ArchiveEntry
		entries, err = sevenzip.ReadEntries(downloadedFile)
		var commonFolderName string
		if err != nil {
			return err
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

		var unzipErr error
		if commonFolderName != "" && build.Location.SuppressParentFolder {
			unzipErr = sevenzip.Uncompress(downloadedFile, os.TempDir())
			if unzipErr == nil {
				sourceFolder := path.Join(os.TempDir(), commonFolderName)
				err = files.MoveFolder(sourceFolder, build.Location.Folder)
			}
		} else {
			unzipErr = sevenzip.Uncompress(downloadedFile, build.Location.Folder)
		}
		if unzipErr == nil {
			os.Remove(downloadedFile)
		} else {
			os.Rename(downloadedFile, downloadedFile+ ".bad")
		}
	}

	if build.PostInstallCmd != "" {
		err = build.ExecutePostBuildCommand(downloadedFile)
		if err != nil {
			return err
		}
		os.Remove(downloadedFile)
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

	return err
}

func (build Build) ExecutePostBuildCommand(localFile string) (err error) {
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
	log.Println("Executing", commandLine)
	combinedOutputBytes, err = exec.Command(commandLine[0], commandLine[1:]...).CombinedOutput()
	if err != nil {
		log.Println("Error occurred : ", strings.TrimSpace(string(combinedOutputBytes)), err)
	}
	return err
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
