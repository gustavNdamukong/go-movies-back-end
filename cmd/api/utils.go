package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// NOTES: 'omitempty' tells Go to omit the Data key in the json file if no data is specified for it
type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// func to write a json data
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// check if any headers are passed in
	if len(headers) > 0 {
		for key, value := range headers[0] {
			// build the response headers with the provided headers
			w.Header()[key] = value
		}
	}

	// set the content type & return the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// We will call this function everytime we want to decode json
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	// we dont want to receive any json file that's bigger than 1 MB
	maxBytes := 1024 * 1024 // one megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	// dont allow fields that we dont know about
	dec.DisallowUnknownFields()

	// decode the data
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	// there should only ever be one json file in the request body
	err = dec.Decode(&struct{}{})
	// if the error is anything other than 'io.EOF' then we have more than one json file
	if err != io.EOF {
		return errors.New("body must only contain a single json value")
	}

	return nil
}

// generate nicely formatted error responses as JSON
func (app *application) errorJSON(w http.ResponseWriter, error error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	payload.Message = error.Error()

	return app.writeJSON(w, statusCode, payload)
}
