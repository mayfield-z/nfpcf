package processor

import (
	"github.com/free5gc/nfpcf/internal/cache"
	"github.com/free5gc/nfpcf/internal/sbi/consumer"
)

type Processor struct {
	cache     *cache.NFProfileCache
	nrfClient *consumer.NRFClient
}

func NewProcessor(cache *cache.NFProfileCache, nrfClient *consumer.NRFClient) *Processor {
	return &Processor{
		cache:     cache,
		nrfClient: nrfClient,
	}
}
