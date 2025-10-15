package validation

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct validates a struct using the validator
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}
