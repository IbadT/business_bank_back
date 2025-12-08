// pkg/validator/validator.go
package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	typ := val.Type()
	
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// Проверка required
		if tag := fieldType.Tag.Get("validate"); strings.Contains(tag, "required") {
			if isZero(field) {
				return ValidationError{
					Field: fieldType.Name,
					Message: "field is required",
				}
			}
		}
		
		// Проверка min/max для чисел
		if field.Kind() == reflect.Float64 || field.Kind() == reflect.Int {
			if err := v.validateMinMax(field, fieldType); err != nil {
				return err
			}
		}
		
		// Проверка regexp
		if tag := fieldType.Tag.Get("validate"); strings.Contains(tag, "regexp=") {
			regexStr := strings.Split(tag, "regexp=")[1]
			if err := v.validateRegex(field, regexStr); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// isZero проверяет, является ли значение нулевым
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	default:
		return false
	}
}

// validateMinMax проверяет min/max значения
func (v *Validator) validateMinMax(field reflect.Value, fieldType reflect.StructField) error {
	tag := fieldType.Tag.Get("validate")
	if tag == "" {
		return nil
	}
	
	var minVal, maxVal *float64
	
	// Парсим min
	if strings.Contains(tag, "min=") {
		minStr := strings.Split(strings.Split(tag, "min=")[1], ",")[0]
		minStr = strings.Split(minStr, " ")[0]
		if val, err := strconv.ParseFloat(minStr, 64); err == nil {
			minVal = &val
		}
	}
	
	// Парсим max
	if strings.Contains(tag, "max=") {
		maxStr := strings.Split(strings.Split(tag, "max=")[1], ",")[0]
		maxStr = strings.Split(maxStr, " ")[0]
		if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
			maxVal = &val
		}
	}
	
	// Парсим gt (greater than)
	if strings.Contains(tag, "gt=") {
		gtStr := strings.Split(strings.Split(tag, "gt=")[1], ",")[0]
		gtStr = strings.Split(gtStr, " ")[0]
		if val, err := strconv.ParseFloat(gtStr, 64); err == nil {
			minVal = &val
		}
	}
	
	var fieldVal float64
	if field.Kind() == reflect.Float64 {
		fieldVal = field.Float()
	} else if field.Kind() == reflect.Int {
		fieldVal = float64(field.Int())
	} else {
		return nil
	}
	
	if minVal != nil && fieldVal < *minVal {
		return ValidationError{
			Field: fieldType.Name,
			Message: fmt.Sprintf("value must be >= %v", *minVal),
		}
	}
	
	if maxVal != nil && fieldVal > *maxVal {
		return ValidationError{
			Field: fieldType.Name,
			Message: fmt.Sprintf("value must be <= %v", *maxVal),
		}
	}
	
	return nil
}

// validateRegex проверяет регулярное выражение
func (v *Validator) validateRegex(field reflect.Value, regexStr string) error {
	if field.Kind() != reflect.String {
		return nil
	}
	
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	if !regex.MatchString(field.String()) {
		return ValidationError{
			Field: field.Type().Name(),
			Message: fmt.Sprintf("value does not match pattern %s", regexStr),
		}
	}
	
	return nil
}

type ValidationError struct {
	Field string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}