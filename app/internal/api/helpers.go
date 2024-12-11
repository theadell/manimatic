package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxBodySize = 1_024 * 1_024

type envelope map[string]any

func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(dst)
	if err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {

		case errors.As(err, &unmarshalTypeError):
			return errors.New("incorrect JSON type for field " + unmarshalTypeError.Field)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("unexpected end of JSON input")

		case errors.Is(err, io.EOF):
			return errors.New("request body cannot be empty")

		default:
			return errors.New("invalid JSON provided")
		}
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func (api *App) errorResponse(w http.ResponseWriter, status int, message any) {
	env := envelope{"error": message}
	err := WriteJSON(w, status, env)
	if err != nil {
		api.logger.Error("failed to write JSON error response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (api *App) badRequestResponse(w http.ResponseWriter, message any) {
	api.errorResponse(w, http.StatusBadRequest, message)
}

func (api *App) serverError(w http.ResponseWriter, err error, args ...any) {
	api.logger.Error(err.Error(), args...)
	message := "the server could not process your request"
	api.errorResponse(w, http.StatusBadRequest, message)
}
