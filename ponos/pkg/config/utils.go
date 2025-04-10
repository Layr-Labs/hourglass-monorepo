package config

import "strings"

func KebabToSnakeCase(str string) string {
	return strings.ReplaceAll(str, "-", "_")
}
