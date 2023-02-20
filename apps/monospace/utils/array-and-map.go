package utils

import (
	"strings"
)

func MapGetKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	i := 0
	for k := range m {
		keys = append(keys, k)
		i++
	}
	return keys
}

func Filter[T any](array []T, predicate func(T) bool) (res []T) {
	for _, val := range array {
		if predicate(val) {
			res = append(res, val)
		}
	}
	return
}

func Map[T any, V any, Mapper func(T) V](array []T, mapper Mapper) (res []V) {
	for _, val := range array {
		res = append(res, mapper(val))
	}
	return
}

func MapAndFilter[T any, V any, Mapper func(T) (V, bool)](array []T, mapper Mapper) (res []V) {
	for _, val := range array {
		newVal, keep := mapper(val)
		if keep {
			res = append(res, newVal)
		}
	}
	return
}

func PrefixPredicate(prefix string) func(string) bool {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix)
	}
}
