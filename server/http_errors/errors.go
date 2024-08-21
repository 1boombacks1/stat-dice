package httpErrors

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"
)

const (
	AppErrTmplName  = "error-block"
	AuthErrTmplName = "auth-err"
)

var errTmpl *template.Template

type ErrResponse struct {
	Title       string `json:"-"`
	Explanation string `json:"error"`
	ElementID   string `json:"-"`

	err        error
	statusCode int
}

func (e *ErrResponse) Error() string {
	return e.err.Error()
}

func (e *ErrResponse) Unwrap() error {
	return e.err
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.statusCode)
	return nil
}

func (e *ErrResponse) WithExplanation(explanation string) *ErrResponse {
	e.Explanation = explanation
	return e
}

func (e *ErrResponse) WithLog(logger *zerolog.Event) *ErrResponse {
	logger.Err(e.err).Send()
	return e
}

func (e *ErrResponse) SetElementID(id string) *ErrResponse {
	e.ElementID = id
	return e
}

func (e *ErrResponse) SetTitle(title string) *ErrResponse {
	e.Title = title
	return e
}

func (e *ErrResponse) Execute(w http.ResponseWriter, templateName string, logger *zerolog.Event) {
	w.WriteHeader(e.statusCode)

	if err := errTmpl.ExecuteTemplate(w, templateName, e); err != nil {
		logger.Err(err).Msg("failed to render auth-err")
	}
}

func NewAuthError(err error, status int) *ErrResponse {
	return &ErrResponse{
		err:         err,
		statusCode:  status,
		Explanation: err.Error(),
	}
}

func ErrUnsupportedMediaType() *ErrResponse {
	return NewAuthError(errors.New("unsupport media type"), http.StatusUnsupportedMediaType)
}

func ErrInternalServer(err error) *ErrResponse {
	return NewAuthError(err, http.StatusInternalServerError)
}

func ErrUnauthorized(err error) *ErrResponse {
	return NewAuthError(err, http.StatusUnauthorized)
}

func ErrBadRequest(err error) *ErrResponse {
	return NewAuthError(err, http.StatusBadRequest)
}

func HTTPNotFound(err error) *ErrResponse {
	return NewAuthError(err, http.StatusNotFound)
}

func HTTPMethodNotAllowed(err error) *ErrResponse {
	return NewAuthError(err, http.StatusMethodNotAllowed)
}

func init() {
	errTmpl = template.Must(template.ParseFS(templates.Err, "errors/*.html"))
}
