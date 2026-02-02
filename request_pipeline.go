package rate_limiter

type requestPipeline struct {
	rateLimiter   iRateLimiter
	trafficShaper iTrafficShapeAlgorithm
}

func newRequestPipeline(rateLimiter iRateLimiter, trafficShaper iTrafficShapeAlgorithm) requestPipeline {
	return requestPipeline{
		rateLimiter:   rateLimiter,
		trafficShaper: trafficShaper,
	}
}

func (r *requestPipeline) handleRequest() RequestPipelineResponse {
	limiter := r.rateLimiter.eval()
	if !<-limiter.Allowed() {
		return newSyncRequestPipelineResponse(false)
	}
	if r.trafficShaper == nil {
		return limiter
	}
	responseChan := r.trafficShaper.addRequest()
	return newAsyncRequestPipelineResponse(responseChan)
}
