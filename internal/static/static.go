package static

import "embed"

//go:embed *.html *.png
var Fs embed.FS
