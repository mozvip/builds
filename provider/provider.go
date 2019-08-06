package provider

import (
	"github.com/mozvip/builds/builds"
	"github.com/mozvip/builds/search"
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
	Update()
	CanHandle(buildType string) bool
	Search(packageName string) []search.Result
	NeedsInstallLocation() bool
	DownloadBuild(providerData *builds.ProviderData, currentVersion *version.Version) search.Result

}

func CheckBuild(build *builds.Build, buildProviders []BuildProvider, currentVersion *version.Version) (version.Version, error) {

	log.Printf("Checking for new build of %s, current version = %s", build.Name, currentVersion)

	for _, prov := range buildProviders {
		if prov.CanHandle(build.Provider.Type) {
			if prov.NeedsInstallLocation() && build.Location.Folder == "" && build.PostInstallCmd == "" {
				log.Printf("%s has not location or post install cmd set, skipping installation", build.Name)
				return *currentVersion, nil
			}
			result := prov.DownloadBuild(&build.Provider, currentVersion)
			if result.Err == nil {
				if result.RemoteURL != "" {
					result.Err = builds.DownloadBuildFromURL(build, result.RemoteURL)
				}
			}

			return result.Version, result.Err
		}
	}

	return *currentVersion, nil
}

