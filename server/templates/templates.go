package templates

import (
	"embed"
	"errors"
	"html/template"
)

//go:embed auth
var Auth embed.FS

//go:embed main
var Main embed.FS

//go:embed errors
var Err embed.FS

type PageContent uint8

const (
	FIND_LOBBY_CONTENT        PageContent = iota // fine-lobbies.html
	CREATE_LOBBY_CONTENT                         // create.html
	LEADERBOARD_CONTENT                          // leaderboard.html
	COMPLETED_LOBBIES_CONTENT                    // completed.html

	LOBBY_CONTENT // lobby.html
)

// Unmarshal only main page for server
func (m *PageContent) UnmarshalText(text []byte) error {
	switch string(text) {
	case "find":
		*m = FIND_LOBBY_CONTENT
	case "create":
		*m = CREATE_LOBBY_CONTENT
	case "completed":
		*m = COMPLETED_LOBBIES_CONTENT
	case "leaderboard":
		*m = LEADERBOARD_CONTENT
	default:
		return errors.New("unknown page type: " + string(text))
	}
	return nil
}

func (m PageContent) String() string {
	return []string{"find", "create", "leaderboard", "completed", "lobby"}[m]
}

func (m PageContent) Filename() string {
	return []string{"find-lobbies.html", "create.html", "leaderboard.html", "completed.html", "lobby.html"}[m]
}

func (m PageContent) GetTemplate(funcs *template.FuncMap) (*template.Template, error) {
	path := "main/sections/" + m.Filename()
	tmpl := template.New(m.Filename())
	if funcs != nil {
		tmpl = tmpl.Funcs(*funcs)
	}
	return tmpl.ParseFS(Main, path)
}
