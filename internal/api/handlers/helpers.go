package handlers

import (
	"errors"
	"reflect"
	"strings"

	"github.com/greatdaveo/Schoolly/pkg/utils"
)

func CheckBlankFields(value interface{}) error {
	val := reflect.ValueOf(value)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			// http.Error(w, "❌ All fields are required", http.StatusBadRequest)
			return utils.ErrorHandler(errors.New("all fields are required"), "❌ All fields are required")
		}
	}
	return nil
}

func GetFieldsName(model interface{}) []string {
	val := reflect.TypeOf(model)
	fields := []string{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldToAdd := strings.TrimSuffix(field.Tag.Get("json"), ",omitempty")
		fields = append(fields, fieldToAdd) // To GET JSON tag
	}
	return fields
}
