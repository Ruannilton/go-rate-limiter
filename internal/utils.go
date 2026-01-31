package internal

import (


	"golang.org/x/exp/constraints"
)



func getNumberFromMap[T constraints.Integer | constraints.Float](m map[string]any, key string) (T, bool) {
	var value T
	switch v := m[key].(type) {
	case int:
		value = T(int(v))
	case float64:
		value = T(float64(v))
	default:
		return T(0), false
	}
	return value, true
}