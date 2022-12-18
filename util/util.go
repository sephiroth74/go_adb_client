package util

import "errors"

func Map[I any, O any](data []I, f func(I) (O, error)) ([]O, error) {
	mapped := make([]O, len(data))
	for i, e := range data {
		m, err := f(e)
		if err != nil {
			return nil, err
		}
		mapped[i] = m
	}
	return mapped, nil
}

func MapNotNull[I any, O any](data []I, f func(I) (O, error)) []O {
	var mapped []O
	for _, e := range data {
		m, err := f(e)
		if err == nil {
			mapped = append(mapped, m)
		}
	}
	return mapped
}

func Any[T any](data []T, f func(T) bool) bool {
	for _, e := range data {
		if f(e) {
			return true
		}
	}
	return false
}

func All[T any](data []T, f func(T) bool) bool {
	for _, e := range data {
		if !f(e) {
			return false
		}
	}
	return true
}

func First[T any](data []T, f func(T) (bool, error)) (*T, error) {
	for _, e := range data {
		r, err := f(e)
		if err != nil {
			return nil, err
		}

		if r {
			return &e, nil
		}
	}
	return nil, errors.New("not found")
}

func FirstNotNull[T any](data []T, f func(T) bool) *T {
	for _, e := range data {
		if f(e) {
			return &e
		}
	}
	return nil
}
