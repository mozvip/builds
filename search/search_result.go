package search

import "github.com/mozvip/builds/version"

type SearchResult struct {
	Label string
	version version.Version
}

func New(label string, version version.Version) SearchResult {
	return SearchResult{Label:label, version:version}
}