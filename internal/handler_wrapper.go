package internal

type HandlerWrapper[T Algorithm] struct {
	storage StorageAlgorithmData[T]
}

func (h *HandlerWrapper[T]) Handle(identifier string) (AlgorithmResponse, error) {
	algoFound, exists := h.storage.Retrieve(identifier)

	if !exists {

		algoCreated, err := h.storage.New(identifier)
		if err != nil {
			return AlgorithmResponse{}, err
		}
		if err := h.storage.Store(identifier, algoCreated); err != nil {
			return AlgorithmResponse{}, err
		}
		algoFound = algoCreated
	}

	return algoFound.Eval(), nil
}

func NewHandler[T Algorithm](storage StorageAlgorithmData[T]) *HandlerWrapper[T] {
	return &HandlerWrapper[T]{
		storage: storage,
	}
}
