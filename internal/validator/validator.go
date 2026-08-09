package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	validate   *validator.Validate
	emailRegex *regexp.Regexp
)

func init() {
	validate = validator.New()
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}
