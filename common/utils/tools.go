package utils

func Contains[T string | int](list []T, target T) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
