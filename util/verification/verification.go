package verification

import (
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"
	"network/model/department"
	"reflect"
	"regexp"
)

var customVerification = map[string]validator.Func{
	"phone":          PhoneValidator,
	"departmentType": DepartmentTypeValidator,
}

func Register() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		for name, ver := range customVerification {
			if err := v.RegisterValidation(name, ver); err != nil {
				return err
			}
		}
	}
}

func PhoneValidator(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	if data, ok := field.Interface().(string); ok {
		str := `^1[3-9]\d{9}$`
		reg := regexp.MustCompile(str)
		if !reg.MatchString(data) {
			return false
		}
	}

	return true
}

func DepartmentTypeValidator(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	if data, ok := field.Interface().(int); ok {
		if data <= department.TypeAdministrator || data >= department.TypeMax {
			return false
		}
	}

	return true
}