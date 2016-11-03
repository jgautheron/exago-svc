package loader

import (
	"encoding/json"
	"fmt"

	"github.com/hotolab/exago-svc/repository/model"
)

// Save persists in database the repository data.
func (l Loader) Save(repo model.Record) error {
	b, err := json.Marshal(repo.GetData())
	if err != nil {
		return err
	}
	return l.config.DB.Put(l.getCacheKey(repo.GetName(), repo.GetBranch()), b)
}

// ClearCache removes the repository from database.
func (l Loader) ClearCache(repo, branch string) error {
	return l.config.DB.Delete(l.getCacheKey(repo, branch))
}

// IsCached checks if the repository's data is cached in database.
func (l Loader) IsCached(repo, branch string) bool {
	if _, err := l.config.DB.Get(l.getCacheKey(repo, branch)); err != nil {
		return false
	}
	return true
}

// cacheKey returns the standardised key format.
func (l Loader) getCacheKey(repo, branch string) []byte {
	return []byte(fmt.Sprintf("%s-%s", repo, branch))
}