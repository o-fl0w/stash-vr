package static

import "embed"

//go:embed *.gohtml *.html *.png
var Fs embed.FS
