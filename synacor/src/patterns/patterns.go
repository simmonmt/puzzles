package patterns

import "regexp"

const (
	NameWithOptionalOffset = `(\w+)(?:\+(\d+))?`
)

var (
	NameWithOptionalOffsetPattern = regexp.MustCompile(`^` + NameWithOptionalOffset + `$`)
)
