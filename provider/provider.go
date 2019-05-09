package provider

import "github.com/mozvip/builds/version"

type Provider struct {
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


type BuildProvider interface {

	DownloadBuild(currentVersion version.Version)

}
