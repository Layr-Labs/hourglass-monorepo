package config

import "strings"

func KebabToSnakeCase(str string) string {
	return strings.ReplaceAll(str, "-", "_")
}

func NormalizeFlagName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func DefaultInt(v, d int) int {
	if v == 0 {
		return d
	}
	return v
}
