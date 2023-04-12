package functional_map

// This library provides functional programming helpers
// whose OUTPUTs are maps
// Reference: HHVM's Dict library
// https://docs.hhvm.com/hsl/reference/function/HH.Lib.Dict.associate/

// FromValues converts a slice to a map using keys
// given by `key_fn` applied to each item.
// If keys are duplicated, latter value will override the previous one.
func FromValues[K comparable, V any](values []V, key_fn func(V) K) map[K]V {
	result := make(map[K]V, len(values))
	for _, item := range values {
		result[key_fn(item)] = item
	}

	return result
}
