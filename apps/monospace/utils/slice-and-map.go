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

func Filter[T any](slice []T, predicate func(T) bool) (res []T) {
	for _, val := range slice {
		if predicate(val) {
			res = append(res, val)
		}
	}
	return
}

func Map[T any, V any, Mapper func(T) V](slice []T, mapper Mapper) []V {
	res := make([]V, len(slice))
	for key, val := range slice {
		res[key] = mapper(val)
	}
	return res
}

func MapAndFilter[T any, V any, Mapper func(T) (V, bool)](slice []T, mapper Mapper) (res []V) {
	for _, val := range slice {
		newVal, keep := mapper(val)
		if keep {
			res = append(res, newVal)
		}
	}
	return
}

func Contains[T string | int](slice []T, search T) bool {
	for _, val := range slice {
		if val == search {
			return true
		}
	}
	return false
}

func PrefixPredicate(prefix string) func(string) bool {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix)
	}
}
