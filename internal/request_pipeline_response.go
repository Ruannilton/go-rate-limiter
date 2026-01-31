package internal

type RequestPipelineResponse struct {
	allowed           bool
	asyncResponse     bool
	asyncResponseChan <-chan bool
}

func NewSyncRequestPipelineResponse(allowed bool) RequestPipelineResponse {
	return RequestPipelineResponse{
		allowed:       allowed,
		asyncResponse: false,
	}
}

func NewAsyncRequestPipelineResponse(asyncResponseChan <-chan bool) RequestPipelineResponse {
	return RequestPipelineResponse{
		asyncResponse:     true,
		asyncResponseChan: asyncResponseChan,
	}
}

func (r *RequestPipelineResponse) Allowed() <-chan bool {
	if r.asyncResponse {
		return r.asyncResponseChan
	}

	ch := make(chan bool, 2)
	ch <- r.allowed
	ch <- r.allowed
	close(ch)
	return ch
}

func (r *RequestPipelineResponse) IsAsync() bool {
	return r.asyncResponse
}