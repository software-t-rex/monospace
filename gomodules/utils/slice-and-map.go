package utils

import (
	"strings"
)

func MapGetKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func SliceFilter[T any](slice []T, predicate func(T) bool) (res []T) {
	for _, val := range slice {
		if predicate(val) {
			res = append(res, val)
		}
	}
	return
}

func MapFilter[K string | int, T any](m map[K]T, predicate func(T) bool) map[K]T {
	res := make(map[K]T, len(m))
	for k, val := range m {
		if predicate(val) {
			res[k] = val
		}
	}
	return res
}

func SliceMap[T any, V any, Mapper func(T) V](slice []T, mapper Mapper) []V {
	res := make([]V, len(slice))
	for key, val := range slice {
		res[key] = mapper(val)
	}
	return res
}

func SliceMapAndFilter[T any, V any, Mapper func(T) (V, bool)](slice []T, mapper Mapper) (res []V) {
	for _, val := range slice {
		newVal, keep := mapper(val)
		if keep {
			res = append(res, newVal)
		}
	}
	return
}

// return "" if not found
func StringMapFind[V string | int](m map[string]V, val V) string {
	for k, v := range m {
		if v == val {
			return k
		}
	}
	return ""
}

// returns -1 if not found
func IntMapFind[V string | int](m map[int]V, val V) int {
	for k, v := range m {
		if v == val {
			return k
		}
	}
	return -1
}

func SliceContains[T string | int](slice []T, search T) bool {
	for _, val := range slice {
		if val == search {
			return true
		}
	}
	return false
}

// return -1 if not found
func SliceFindIndex[T string | int](slice []T, search T) int {
	for k, v := range slice {
		if v == search {
			return k
		}
	}
	return -1
}

func SliceReverse[T any](slice []T) []T {
	a := make([]T, len(slice))
	copy(a, slice)
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}

func PrefixPredicate(prefix string) func(string) bool {
	return func(s string) bool {
		return strings.HasPrefix(s, prefix)
	}
}
