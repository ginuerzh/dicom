package tag

var (
	customDict map[Tag]TagInfo
)

// SetCustomDict sets the custom dictionary.
func SetCustomDict(dict map[Tag]TagInfo) {
	customDict = dict
}
