package cache

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/free5gc/openapi/models"
)

type CacheEntry struct {
	Profile   *models.NrfNfDiscoveryNfProfile
	ExpiresAt time.Time
}

type SearchResultEntry struct {
	Result    *models.SearchResult
	ExpiresAt time.Time
}

type NFProfileCache struct {
	profiles       map[string]*CacheEntry
	typeIndex      map[string][]string
	searchResults  map[string]*SearchResultEntry
	lock           sync.RWMutex
	defaultTTL     time.Duration
	cleanupTimer   *time.Ticker
}

func NewNFProfileCache(ttl time.Duration) *NFProfileCache {
	cache := &NFProfileCache{
		profiles:      make(map[string]*CacheEntry),
		typeIndex:     make(map[string][]string),
		searchResults: make(map[string]*SearchResultEntry),
		defaultTTL:    ttl,
		cleanupTimer:  time.NewTicker(ttl / 2),
	}

	go cache.cleanupExpired()
	return cache
}

func (c *NFProfileCache) Put(profile *models.NrfNfDiscoveryNfProfile) {
	c.lock.Lock()
	defer c.lock.Unlock()

	nfInstanceID := profile.NfInstanceId
	entry := &CacheEntry{
		Profile:   profile,
		ExpiresAt: time.Now().Add(c.defaultTTL),
	}

	c.profiles[nfInstanceID] = entry

	if profile.NfType != "" {
		c.addToTypeIndex(string(profile.NfType), nfInstanceID)
	}
}

func (c *NFProfileCache) Get(nfInstanceID string) (*models.NrfNfDiscoveryNfProfile, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	entry, exists := c.profiles[nfInstanceID]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Profile, true
}

func (c *NFProfileCache) Delete(nfInstanceID string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if entry, exists := c.profiles[nfInstanceID]; exists {
		if entry.Profile.NfType != "" {
			c.removeFromTypeIndex(string(entry.Profile.NfType), nfInstanceID)
		}
		delete(c.profiles, nfInstanceID)
	}
}

func (c *NFProfileCache) Search(queryParams url.Values) []*models.NrfNfDiscoveryNfProfile {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var results []*models.NrfNfDiscoveryNfProfile

	targetNfType := queryParams.Get("target-nf-type")
	if targetNfType == "" {
		return results
	}

	instanceIDs, exists := c.typeIndex[targetNfType]
	if !exists {
		return results
	}

	now := time.Now()
	for _, id := range instanceIDs {
		entry, exists := c.profiles[id]
		if !exists || now.After(entry.ExpiresAt) {
			continue
		}

		if c.matchesQuery(entry.Profile, queryParams) {
			results = append(results, entry.Profile)
		}
	}

	return results
}

func (c *NFProfileCache) matchesQuery(
	profile *models.NrfNfDiscoveryNfProfile,
	queryParams url.Values,
) bool {
	if requesterNfType := queryParams.Get("requester-nf-type"); requesterNfType != "" {
		if !c.matchesNfType(profile, requesterNfType) {
			return false
		}
	}

	if snssais := queryParams["snssais"]; len(snssais) > 0 {
		if !c.matchesSnssais(profile, snssais) {
			return false
		}
	}

	if dnn := queryParams.Get("dnn"); dnn != "" {
		if !c.matchesDnn(profile, dnn) {
			return false
		}
	}

	return true
}

func (c *NFProfileCache) matchesNfType(
	profile *models.NrfNfDiscoveryNfProfile,
	requesterNfType string,
) bool {
	return true
}

func (c *NFProfileCache) matchesSnssais(
	profile *models.NrfNfDiscoveryNfProfile,
	querySnssais []string,
) bool {
	if profile.SNssais == nil || len(profile.SNssais) == 0 {
		return true
	}

	for _, querySnssai := range querySnssais {
		for _, profileSnssai := range profile.SNssais {
			if c.snssaiEquals(&profileSnssai, querySnssai) {
				return true
			}
		}
	}
	return false
}

func (c *NFProfileCache) matchesDnn(profile *models.NrfNfDiscoveryNfProfile, queryDnn string) bool {
	if profile.SmfInfo == nil || profile.SmfInfo.SNssaiSmfInfoList == nil {
		return true
	}

	for _, smfInfo := range profile.SmfInfo.SNssaiSmfInfoList {
		if smfInfo.DnnSmfInfoList != nil {
			for _, dnnInfo := range smfInfo.DnnSmfInfoList {
				if dnnInfo.Dnn == queryDnn {
					return true
				}
			}
		}
	}
	return false
}

func (c *NFProfileCache) snssaiEquals(
	profileSnssai *models.ExtSnssai,
	querySnssai string,
) bool {
	return true
}

func (c *NFProfileCache) addToTypeIndex(nfType string, nfInstanceID string) {
	if c.typeIndex[nfType] == nil {
		c.typeIndex[nfType] = make([]string, 0)
	}

	for _, id := range c.typeIndex[nfType] {
		if id == nfInstanceID {
			return
		}
	}

	c.typeIndex[nfType] = append(c.typeIndex[nfType], nfInstanceID)
}

func (c *NFProfileCache) removeFromTypeIndex(nfType string, nfInstanceID string) {
	ids := c.typeIndex[nfType]
	for i, id := range ids {
		if id == nfInstanceID {
			c.typeIndex[nfType] = append(ids[:i], ids[i+1:]...)
			return
		}
	}
}

func (c *NFProfileCache) GetSearchResult(queryParams url.Values) (*models.SearchResult, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	key := generateSearchKey(queryParams)
	entry, exists := c.searchResults[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Result, true
}

func (c *NFProfileCache) SetSearchResult(queryParams url.Values, result *models.SearchResult) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := generateSearchKey(queryParams)
	entry := &SearchResultEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(c.defaultTTL),
	}
	c.searchResults[key] = entry
}

func generateSearchKey(queryParams url.Values) string {
	targetNfType := queryParams.Get("target-nf-type")
	requesterNfType := queryParams.Get("requester-nf-type")
	return fmt.Sprintf("%s:%s", targetNfType, requesterNfType)
}

func (c *NFProfileCache) cleanupExpired() {
	for range c.cleanupTimer.C {
		c.lock.Lock()
		now := time.Now()
		for id, entry := range c.profiles {
			if now.After(entry.ExpiresAt) {
				if entry.Profile.NfType != "" {
					c.removeFromTypeIndex(string(entry.Profile.NfType), id)
				}
				delete(c.profiles, id)
			}
		}
		for key, entry := range c.searchResults {
			if now.After(entry.ExpiresAt) {
				delete(c.searchResults, key)
			}
		}
		c.lock.Unlock()
	}
}

func (c *NFProfileCache) Stop() {
	if c.cleanupTimer != nil {
		c.cleanupTimer.Stop()
	}
}
