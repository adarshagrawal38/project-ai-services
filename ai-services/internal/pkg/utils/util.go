package utils

import (
	"strings"
	"unicode"
)

// CapitalizeAndFormat replaces '_' and '-' with spaces, then capitalizes each word.
func CapitalizeAndFormat(s string) string {
	// Replace _ and - with spaces
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Split into words
	words := strings.Fields(s)

	// Capitalize each word
	for i, w := range words {
		if len(w) > 0 {
			runes := []rune(w)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}

	// Join back into a single string
	return strings.Join(words, " ")
}

// BoolPtr -> converts to bool ptr
func BoolPtr(v bool) *bool {
	return &v
}

// flattenArray takes a 2D slice and returns a 1D slice with all values
func FlattenArray[T interface{ ~int | ~float64 | ~string }](arr [][]T) []T {
	flatArr := []T{}

	for _, row := range arr {
		flatArr = append(flatArr, row...)
	}
	return flatArr
}
