package util

func Map[T interface{}](data []string, f func(string) (T, error)) ([]T, error) {
	mapped := make([]T, len(data))
	for i, e := range data {
		m, err := f(e)
		if err != nil {
			return nil, err
		}
		mapped[i] = m
	}
	return mapped, nil
}

func MapNotNull[T interface{}](data []string, f func(string) (T, error)) []T {
	mapped := []T{}
	for _, e := range data {
		m, err := f(e)
		if err == nil {
			mapped = append(mapped, m)
		}
	}
	return mapped
}

func Any[T interface{}](data []T, f func(T) bool) bool {
	for _, e := range data {
		if f(e) {
			return true
		}
	}
	return false
}

func All[T interface{}](data []T, f func(T) bool) bool {
	for _, e := range data {
		if !f(e) {
			return false
		}
	}
	return true
}

func First[T interface{}](data []T, f func(T) bool) *T {
	for _, e := range data {
		if f(e) {
			return &e
		}
	}
	return nil
}
