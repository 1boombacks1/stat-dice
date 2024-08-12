package templates

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed auth/*.html
var authFS embed.FS

//go:embed main/*.html
var mainFS embed.FS

var (
	authTmpl *template.Template
	mainTmpl *template.Template
)

func ExecuteAuth(w http.ResponseWriter) error {
	return authTmpl.ExecuteTemplate(w, "auth", nil)
}

func WriteAuthError(w http.ResponseWriter, elementID string, status int, err error) {
	w.WriteHeader(status)

	if err := authTmpl.ExecuteTemplate(w, "auth-err",
		struct {
			ElementID string
			Error     string
		}{
			ElementID: elementID,
			Error:     err.Error(),
		}); err != nil {
		panic("failed to EXECUTE auth-err-template: " + err.Error())
	}
}

func init() {
	authTmpl = template.Must(template.ParseFS(authFS, "*.html"))
	mainTmpl = template.Must(template.ParseFS(mainFS, "*.html"))
}
