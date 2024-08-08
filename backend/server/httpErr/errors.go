package httpErr

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	Ok          bool   `json:"ok"`
	Explanation string `json:"error,omitempty"`

	err        error `json:"-"`
	statusCode int   `json:"-"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.statusCode)
	render.DefaultResponder(w, r, e)
	return nil
}

func (e *ErrResponse) Error() string {
	return e.err.Error()
}

func (e *ErrResponse) Unwrapp() error {
	return e.err
}

func (e *ErrResponse) WithExplanation(explanation string) *ErrResponse {
	e.Explanation = explanation
	return e
}

func makeHTTPError(baseErr error, status int) *ErrResponse {
	httpParentErr := &ErrResponse{}
	if errors.Is(baseErr, httpParentErr) {
		// rebase to prevent same-type error wrapping
		baseErr = httpParentErr.err
	}

	return &ErrResponse{
		err:        baseErr,
		statusCode: status,
		Ok:         false,
	}
}

func NewHTTPError(err error, status int) *ErrResponse {
	return makeHTTPError(err, status)
}

func NewHTTPErrorWithExplanation(err error, status int, expl string) *ErrResponse {
	return makeHTTPError(err, status).WithExplanation(expl)
}

func HTTPInternalServerError(err error) *ErrResponse {
	return makeHTTPError(err, http.StatusInternalServerError).WithExplanation("internal server error")
}

func HTTPBadRequest(err error) *ErrResponse {
	return makeHTTPError(err, http.StatusInternalServerError).WithExplanation("bad request")
}

func HTTPNotFound(err error) *ErrResponse {
	return makeHTTPError(err, http.StatusNotFound).WithExplanation("page not found")
}

func HTTPMethodNotAllowed(err error) *ErrResponse {
	return makeHTTPError(err, http.StatusMethodNotAllowed).WithExplanation("method not allowed")
}

func HTTPUnsupportedMediaType(err error) *ErrResponse {
	return makeHTTPError(err, http.StatusUnsupportedMediaType).WithExplanation("only 'application/json' media type is supported")
}

func HTTPUnauthorized(err error) *ErrResponse {
	return makeHTTPError(err, http.StatusUnauthorized).WithExplanation("unauthorized")
}
