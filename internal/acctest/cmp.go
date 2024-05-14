package acctest

import (
	"fmt"

	"github.com/oapi-codegen/nullable"
)

func EqualField[T comparable](field string, a, b T) error {
	if err := Equal(a, b); err != nil {
		return fmt.Errorf(`field "%s": %s`, field, err.Error())
	}

	return nil
}

func Equal[T comparable](a, b T) error {
	if a == b {
		return nil
	}

	return fmt.Errorf(`expected %#v, got %#v`, b, a)
}

func NotEqual[T comparable](a, b T) error {
	if a != b {
		return nil
	}

	return fmt.Errorf(`expected different value than %#v`, a)
}

func NullField[T any](field string, v nullable.Nullable[T]) error {
	if err := Null(v); err != nil {
		return fmt.Errorf(`field "%s": %s`, field, err.Error())
	}

	return nil
}

func Null[T any](v nullable.Nullable[T]) error {
	if v.IsNull() || !v.IsSpecified() {
		return nil
	}

	return fmt.Errorf(`expected null value, got %#v`, v.MustGet())
}

func NullFieldEqual[T comparable](field string, a nullable.Nullable[T], b T) error {
	if err := NullEqual(a, b); err != nil {
		return fmt.Errorf(`field "%s": %s`, field, err.Error())
	}

	return nil
}

func NullEqual[T comparable](a nullable.Nullable[T], b T) error {
	av, err := a.Get()
	if err != nil {
		if a.IsNull() {
			return fmt.Errorf(`expected value %#v, got null`, b)
		}

		return fmt.Errorf(`expected value %#v, got unspecified`, b)
	}

	if av == b {
		return nil
	}

	return fmt.Errorf(`expected value %#v, got %#v`, b, av)
}
