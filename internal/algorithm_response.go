package internal

type AlgorithmResponse struct {
	Allowed             bool
	Delayed             bool
	DelayedResponseChan chan bool
}

func (self *AlgorithmResponse) IsDelayed() bool {
	return self.Delayed
}

func (self *AlgorithmResponse) IsAccepted() bool {
	return self.Allowed
}

func (self *AlgorithmResponse) IsAllowed() bool {
	if self.Delayed {
		return <-self.DelayedResponseChan
	}
	return self.Allowed
}
