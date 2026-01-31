package internal

type IRateLimiter interface {
	Eval() RequestPipelineResponse
}

type ITrafficShapeAlgorithm interface {
	AddRequest() <-chan bool
}
