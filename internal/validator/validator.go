package validator

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors    map[string]error
	NonFieldErrors []error
}

func NewValidator() *Validator {
	v := &Validator{
		FieldErrors: make(map[string]error),
	}
	return v
}

func (v *Validator) CheckValidity() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	err, ok := v.FieldErrors[key]
	newErr := errors.New(message)
	if ok {
		v.FieldErrors[key] = errors.Join(err, newErr)
	} else {
		v.FieldErrors[key] = newErr
	}
}

// if condition is not true, add new field error with key and message args
func (v *Validator) CheckField(condition bool, key, message string) {
	if !condition {
		v.AddFieldError(key, message)
	}
}

func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, errors.New(message))
}

func StringNotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}

var EmailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func StringMatch(s string, re *regexp.Regexp) bool {
	return re.MatchString(s)
}

// check if the number of runes in the string s equals to the int n
func RunesEqualTo(s string, n int) bool {
	return utf8.RuneCountInString(s) == n
}

// check if the number of runes in the string s is greater than the int n
func RunesGreaterThan(s string, n int) bool {
	return utf8.RuneCountInString(s) > n
}

// check if the number of runes in the string s is less than the int n
func RunesLessThan(s string, n int) bool {
	return utf8.RuneCountInString(s) < n
}

// check if the number of runes in the string s is greater than or equal to the int n
func RunesMin(s string, n int) bool {
	return RunesEqualTo(s, n) || RunesGreaterThan(s, n)
}

// check if the number of runes in the string s is less than or equal to the int n
func RunesMax(s string, n int) bool {
	return RunesLessThan(s, n) || RunesEqualTo(s, n)
}

// check if the value v is one of the permitted values
func CheckPermitted[E comparable](v E, permittedValues ...E) bool {
	return slices.Contains(permittedValues, v)
}
