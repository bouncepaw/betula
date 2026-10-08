// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package bxslices provides slice tricks.
package bxslices

func insertElement[T any](array []T, value T, index int) []T {
	return append(array[:index], append([]T{value}, array[index:]...)...)
}

func removeElement[T any](array []T, index int) []T {
	return append(array[:index], array[index+1:]...)
}

func MoveElement[T any](array []T, srcIndex int, dstIndex int) []T {
	value := array[srcIndex]
	return insertElement(removeElement(array, srcIndex), value, dstIndex)
}

func First[T any](array []T) (T, bool) {
	if len(array) == 0 {
		var t T
		return t, false
	}
	return array[0], true
}

func FirstOrEmpty[T any](array []T) T {
	i, _ := First(array)
	return i
}
