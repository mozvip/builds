package search

import "github.com/mozvip/builds/version"

type Result struct {
	Label string
	Version version.Version
	RemoteURL string
	Err error
}

func New(label string, version version.Version, remoteURL string) Result {
	return Result{Label: label, Version:version, RemoteURL:remoteURL}
}

func Error(err error) Result {
	return Result{Err: err}
}

func Installed(version version.Version) Result {
	return Result{Version:version}
}

func None() Result {
	return Result{}
}