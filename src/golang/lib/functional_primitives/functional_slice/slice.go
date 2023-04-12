package functional_slice

// This library provides functional programming helpers
// whose OUTPUTs are slices
// Reference: HHVM's Vec library
// https://docs.hhvm.com/hsl/reference/function/HH.Lib.Vec.map_with_key/

// Map() converts a slice to a new slice with call_back_fn
// applied to each element.
func Map[V1 any, V2 any](slice []V1, call_back_fn func(V1) V2) []V2 {
	result := make([]V2, 0, len(slice))
	for _, item := range slice {
		result = append(result, call_back_fn(item))
	}

	return result
}
