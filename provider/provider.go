package provider

import (
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/version"
	"log"
)

func Init() []BuildProvider {
	enabledProviders := []BuildProvider{
		&ScoopProvider{}, &ChocolateyProvider{}, &HttpLinkProvider{}, &AppVeyorProvider{}, &GitHubProvider{},
	}
	for _, provider := range enabledProviders {
		provider.Init()
	}

	return enabledProviders
}


type BuildProvider interface {

	Init()
	CanHandle(buildType string) bool
	NeedsInstallLocation() bool
	DownloadBuild(build *builds.Build, currentVersion *version.Version) (version.Version, error)

}

func CheckBuild(build *builds.Build, buildProviders []BuildProvider, currentVersion *version.Version) (version.Version, error) {

	log.Printf("Checking for new build of %s, current version = %s", build.Name, currentVersion)

	var err error
	var newVersion version.Version

	for _, prov := range buildProviders {
		if prov.CanHandle(build.Provider.Type) {
			if prov.NeedsInstallLocation() && build.Location.Folder == "" && build.PostInstallCmd == "" {
				log.Printf("%s has not location or post install cmd set, skipping installation", build.Name)
				return *currentVersion, nil
			}
			newVersion, err = prov.DownloadBuild(build, currentVersion)
			break
		}
	}

	return newVersion, err
}

