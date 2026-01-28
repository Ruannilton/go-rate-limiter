package internal



type AlgorithmType int

const (
	FixedWindow AlgorithmType = iota
	SlidingWindowLog
	LeakyBucket
	TokenBucket
)

type RateLimiterBuilder struct {
	router *Router
}

func NewRateLimiterBuilder() *RateLimiterBuilder {
	limiter  := &RateLimiterBuilder{
		router: NewRouter(),
	}

	return limiter
}

func (b*  RateLimiterBuilder) AddRule(resource string, algorithmType AlgorithmType){
	var handler AlgorithmHandler

	switch algorithmType {
	case FixedWindow:
		handler = NewFixedWindowHandler()
	case SlidingWindowLog:
		handler = NewSlidingWindowLogHandler()
	case LeakyBucket:
		handler = NewLeakyBucketHandler()
	case TokenBucket:
		handler = NewTokenBucketHandler()
	}
	b.router.AddRoute(resource, handler)
}

func NewTokenBucketHandler() AlgorithmHandler {
	tbStorage := NewTokenBucketMemoryStorage()
	tbHandler := NewHandler(tbStorage)
	return tbHandler
}

func NewLeakyBucketHandler() AlgorithmHandler {
	lbStorage := NewLeakyBucketMemoryStorage()
	lbHandler := NewHandler(lbStorage)
	return lbHandler
}

func NewSlidingWindowLogHandler() AlgorithmHandler {
	swlStorage := NewSlidingWindowLogMemoryStorage()
	swlHandler := NewHandler(swlStorage)
	return swlHandler
}

func NewFixedWindowHandler() AlgorithmHandler {
	fwStorage := NewFixedWindowMemoryStorage()
	fwHandler := NewHandler(fwStorage)
	return fwHandler

}