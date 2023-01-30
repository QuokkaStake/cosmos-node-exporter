package constants

import "regexp"

const (
	MetricsPrefix = "cosmos_node_exporter_"
)

var (
	GithubRegexp = regexp.MustCompile("https://github.com/(?P<Org>[a-zA-Z0-9-].*)/(?P<Repo>[a-zA-Z0-9-].*)")
)
