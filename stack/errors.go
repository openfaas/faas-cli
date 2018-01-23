package stack

import "errors"

var (
	ErrorMissingFilterOrRegexFlag = errors.New("no functions matching --filter/--regex were found in the YAML file")
	ErrorExclusiveFilterRegexFlag = errors.New("pass in a regex or a filter, not both")
)
