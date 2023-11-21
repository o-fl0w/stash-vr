package build

var Version = "???"
var SHA = "???"

func FullVersion() string {
	return Version + "+" + SHA
}
