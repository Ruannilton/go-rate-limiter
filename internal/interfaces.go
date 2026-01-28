package internal

import "time"

const EXPIRATION_CHECK_INTERVAL = time.Minute * 5

type StorageAlgorithmData[T Algorithm] interface {
	SetDefaultValue(key string, value any)
	New(key string) (T, error)
	Store(key string, value T) error
	Retrieve(key string) (T, bool)
	Delete(key string) error
}

type Algorithm interface {
	Eval() AlgorithmResponse
}

type AlgorithmHandler interface {
	Handle(identifier string) (AlgorithmResponse, error)
}
