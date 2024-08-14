package templates

import "embed"

//go:embed auth
var Auth embed.FS

//go:embed main
var Main embed.FS
