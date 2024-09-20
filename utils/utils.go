package utils

func Filter[T any](s []T, callback func(T) bool) []T {
	var result []T
	for _, v := range s {
		if callback(v) {
			result = append(result, v)
		}
	}
	return result
}
