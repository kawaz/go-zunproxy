package main

// set by ldflags -X main.xxx=yyy
var (
	version string
	tag     string
	commit  string
	date    string
	builtBy string

	build = buildInfo{
		Version: &version,
		Tag:     &tag,
		Commit:  &commit,
		Date:    &date,
		BuiltBy: &builtBy,
	}
)

type buildInfo struct {
	Version *string
	Tag     *string
	Commit  *string
	Date    *string
	BuiltBy *string
}
