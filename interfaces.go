package rate_limiter

type iRateLimiter interface {
	eval() RequestPipelineResponse
}

type iTrafficShapeAlgorithm interface {
	addRequest() <-chan bool
}
