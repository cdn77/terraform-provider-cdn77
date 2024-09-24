package util

func If[T any](condition bool, ifTrueValue T, elseValue T) T { //nolint:revive // yes, it's a control flag
	if condition {
		return ifTrueValue
	}

	return elseValue
}

func Pointer[T any](v T) *T {
	return &v
}
