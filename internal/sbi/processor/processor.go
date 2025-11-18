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

func (p *Processor) GetCache() *cache.NFProfileCache {
	return p.cache
}

func (p *Processor) GetNRFClient() *consumer.NRFClient {
	return p.nrfClient
}
