package utils

import "slices"

func UniqueList[E comparable](s []E) []E {
	var result []E
	for _, item := range s {
		if !slices.Contains(result, item) {
			result = append(result, item)
		}
	}
	return result
}
