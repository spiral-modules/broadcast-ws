package ws

type rpcService struct {
	svc *Service
}

func (r *rpcService) Warmup(topic string, ok *bool) error {
	return nil
}
