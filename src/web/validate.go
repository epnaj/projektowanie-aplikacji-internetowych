package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"unicode/utf8"
)

// validatable is implemented by request DTOs that can check their own fields.
// Keeping validation as plain Go methods avoids a third-party validation
// library: each request type states its own rules next to its definition.
type validatable interface {
	Validate() error
}

var errEmptyBody = errors.New("empty request body")

// decodeAndValidate parses the JSON body into dst, rejecting unknown fields so
// client typos surface as errors instead of being silently dropped, then runs
// dst.Validate() when the type implements it.
func decodeAndValidate(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errEmptyBody
		}
		return err
	}
	if v, ok := dst.(validatable); ok {
		return v.Validate()
	}
	return nil
}

// validateEmail checks the field is a syntactically valid email address.
func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("email must be a valid address")
	}
	return nil
}

// validateLen checks the rune length of value falls within [min, max].
func validateLen(field, value string, min, max int) error {
	n := utf8.RuneCountInString(value)
	if n < min || n > max {
		return fmt.Errorf("%s must be between %d and %d characters", field, min, max)
	}
	return nil
}
