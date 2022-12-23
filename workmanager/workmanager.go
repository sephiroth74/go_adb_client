package workmanager

import "github.com/sephiroth74/go_adb_client/types"

type WorkManager struct {
}

func (w WorkManager) Execute(works ...Work) chan types.Pair[Data, error] {
	data := Data{}
	dataChannel := make(chan types.Pair[Data, error])

	go func() {
		defer close(dataChannel)

		for _, worker := range works {
			result, err := worker.Execute(data)
			if err != nil {
				result := types.Pair[Data, error]{First: Data{}, Second: err}
				dataChannel <- result
				return
			}

			for k, v := range result {
				data[k] = v
			}
		}
		result := types.Pair[Data, error]{First: data, Second: nil}
		dataChannel <- result
	}()

	return dataChannel
}

type Data map[string]any

type Work interface {
	Execute(inputParams Data) (Data, error)
}
