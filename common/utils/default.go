package utils

// Default s 不为空返回 s, 为空返回 d
func Default(s, d string) string {
	if len(s) == 0 {
		return d
	}
	return s
}
