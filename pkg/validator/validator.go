package validator

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/yourusername/go-gin-template/pkg/response"
)

var (
	once     sync.Once
	validate *validator.Validate
)

// Init initializes the validator with custom validations
func Init() {
	once.Do(func() {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			validate = v

			// Register custom tag name function for JSON field names
			v.RegisterTagNameFunc(func(fld reflect.StructField) string {
				name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
				if name == "-" {
					return ""
				}
				return name
			})

			// Register custom validators
			registerCustomValidators(v)
		}
	})
}

// registerCustomValidators registers custom validation tags
func registerCustomValidators(v *validator.Validate) {
	// Password validation: min 8 chars, at least 1 uppercase, 1 lowercase, 1 number
	_ = v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
		return hasUpper && hasLower && hasNumber
	})

	// Phone number validation (basic)
	_ = v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Basic phone validation: starts with + and contains only digits
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		return phoneRegex.MatchString(phone)
	})

	// Username validation: alphanumeric and underscore, 3-30 chars
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) < 3 || len(username) > 30 {
			return false
		}
		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		return usernameRegex.MatchString(username)
	})

	// Slug validation: lowercase alphanumeric and hyphens
	_ = v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		slug := fl.Field().String()
		slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
		return slugRegex.MatchString(slug)
	})

	// No whitespace validation
	_ = v.RegisterValidation("nowhitespace", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return !strings.Contains(value, " ")
	})

	// Alphanumeric validation
	_ = v.RegisterValidation("alphanumeric", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
		return alphanumericRegex.MatchString(value)
	})
}

// FormatValidationErrors converts validator errors to response format
func FormatValidationErrors(err error) []response.ValidationError {
	var errors []response.ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, response.ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

// getErrorMessage returns a human-readable error message
func getErrorMessage(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + e.Param() + " characters"
	case "max":
		return field + " must be at most " + e.Param() + " characters"
	case "len":
		return field + " must be exactly " + e.Param() + " characters"
	case "gte":
		return field + " must be greater than or equal to " + e.Param()
	case "lte":
		return field + " must be less than or equal to " + e.Param()
	case "gt":
		return field + " must be greater than " + e.Param()
	case "lt":
		return field + " must be less than " + e.Param()
	case "eq":
		return field + " must be equal to " + e.Param()
	case "ne":
		return field + " must not be equal to " + e.Param()
	case "oneof":
		return field + " must be one of: " + e.Param()
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	case "alpha":
		return field + " must contain only letters"
	case "alphanum":
		return field + " must contain only letters and numbers"
	case "numeric":
		return field + " must be numeric"
	case "password":
		return field + " must be at least 8 characters with uppercase, lowercase, and number"
	case "phone":
		return field + " must be a valid phone number"
	case "username":
		return field + " must be 3-30 characters, alphanumeric and underscores only"
	case "slug":
		return field + " must be a valid slug (lowercase, alphanumeric, hyphens)"
	case "nowhitespace":
		return field + " must not contain whitespace"
	case "alphanumeric":
		return field + " must contain only letters and numbers"
	case "eqfield":
		return field + " must match " + e.Param()
	case "nefield":
		return field + " must not match " + e.Param()
	default:
		return field + " failed validation: " + e.Tag()
	}
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) []response.ValidationError {
	Init()

	if validate == nil {
		return nil
	}

	err := validate.Struct(s)
	if err != nil {
		return FormatValidationErrors(err)
	}

	return nil
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	Init()

	if validate == nil {
		return nil
	}

	return validate.Var(field, tag)
}
