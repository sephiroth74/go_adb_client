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