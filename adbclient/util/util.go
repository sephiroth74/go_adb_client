package util


func Map(data []string, f func(string) (byte, error)) ([]byte, error) {
	mapped := make([]byte, len(data))
	for i, e := range data {
		m, err := f(e)
		if err != nil {
			return nil, err
		}
		mapped[i] = m
	}
	return mapped, nil
}