package common

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator"
)

// ValidateInputs use for validate all input from request params
func ValidateInputs(input interface{}) (bool, map[string][]string) {
	var validate *validator.Validate
	var result bool
	validate = validator.New()
	err := validate.Struct(input)

	// Validation syntax invalid
	if err != nil {
		if err, ok := err.(*validator.ValidationErrors); ok {
			panic(err)
		}
	}

	// Validation errors occured
	errors := make(map[string][]string)

	reflected := reflect.ValueOf(input)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field, _ := reflected.Type().FieldByName(err.StructField())
			var name string

			if name = field.Tag.Get("json"); name == "" {
				key := strings.Split(err.StructNamespace(), ".")
				if len(key) != 1 {
					name = strings.ToLower(strings.Join(key[1:], "."))
				} else {
					name = strings.ToLower(err.StructField())
				}

			}

			switch err.Tag() {
			case "required":
				errors[name] = append(errors[name], "This field is required")
				// fmt.Println("VALIDATE INPUT ERROR :: The " + name + " is required")
				break
			case "email":
				errors[name] = append(errors[name], "This field should be a valid email")
				// fmt.Println("VALIDATE INPUT ERROR :: The " + name + " should be a valid email")
				break
			case "eqfield":
				errors[name] = append(errors[name], "The "+name+" should be equal to the "+err.Param())
				// fmt.Println("VALIDATE INPUT ERROR :: The " + name + " should be equal to the " + err.Param())
				break
			case "max":
				errors[name] = append(errors[name], "should not be more than "+err.Param())
				// fmt.Println("VALIDATE INPUT ERROR :: The " + name + " should be equal not to the " + err.Param())
				break
			case "len":
				errors[name] = append(errors[name], "lenght must be "+err.Param())
				// fmt.Println("VALIDATE INPUT ERROR :: lenght of " + name + " must be " + err.Param())
			case "oneof":
				errors[name] = append(errors[name], "value must be one of "+strings.Join(strings.Split(err.Param(), " "), ","))
				// fmt.Println("VALIDATE INPUT ERROR :: value of " + name + " must be one of " + strings.Join(strings.Split(err.Param(), " "), ","))
			default:
				errors[name] = append(errors[name], "This field is invalid")
				// fmt.Println("VALIDATE INPUT ERROR :: The " + name + " is invalid")
				break
			}
		}
	}

	if len(errors) == 0 {
		result = true
	} else {
		result = false
	}

	return result, errors
}
