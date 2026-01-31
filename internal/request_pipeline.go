package internal





type RequestPipeline struct {
	rateLimiter   IRateLimiter
	trafficShaper ITrafficShapeAlgorithm
}


func NewRequestPipeline(rateLimiter IRateLimiter, trafficShaper ITrafficShapeAlgorithm) RequestPipeline {
	return RequestPipeline{
		rateLimiter:   rateLimiter,
		trafficShaper: trafficShaper,
	}
}

func (r *RequestPipeline) HandleRequest() (RequestPipelineResponse, error) {
	limiter := r.rateLimiter.Eval()
	if !<-limiter.Allowed() {
		return NewSyncRequestPipelineResponse(false), nil
	}
	if r.trafficShaper == nil {
		return limiter, nil
	}
	responseChan := r.trafficShaper.AddRequest()
	return NewAsyncRequestPipelineResponse(responseChan), nil
}




