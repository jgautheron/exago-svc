package repository

import (
	"fmt"
	"reflect"
	"time"

	"github.com/exago/svc/leveldb"
	"github.com/exago/svc/repository/model"
)

var (
	// DefaultLinters ran by default in Lambda.
	DefaultLinters = []string{
		"deadcode", "dupl", "errcheck", "goconst", "gocyclo", "gofmt", "goimports",
		"golint", "gosimple", "ineffassign", "staticcheck", "vet", "vetshadow",
	}

	// Make sure it satisfies the interface.
	_ RepositoryData = (*Repository)(nil)
)

type Repository struct {
	Name, Branch string

	// Data types
	CodeStats     model.CodeStats
	Imports       model.Imports
	TestResults   model.TestResults
	LintMessages  model.LintMessages
	Metadata      model.Metadata
	Score         model.Score
	ExecutionTime string
	LastUpdate    time.Time

	StartTime time.Time

	db leveldb.Database
}

func New(repo, branch string) *Repository {
	return &Repository{
		Name:   repo,
		Branch: branch,
		db:     leveldb.GetInstance(),
	}
}

// IsCached checks if the repository's data is cached in database.
func (r *Repository) IsCached() bool {
	prefix := fmt.Sprintf("%s-%s", r.Name, r.Branch)
	data, err := r.db.FindAllForRepository([]byte(prefix))
	if err != nil || len(data) != 8 {
		return false
	}
	return true
}

// IsLoaded checks if the data is already loaded.
func (r *Repository) IsLoaded() bool {
	if r.CodeStats == nil {
		return false
	}
	if r.Imports == nil {
		return false
	}
	if reflect.DeepEqual(r.TestResults, model.TestResults{}) {
		return false
	}
	if r.LintMessages == nil {
		return false
	}
	return true
}

// Load retrieves the entire matching dataset from database.
func (r *Repository) Load() (err error) {
	if _, err = r.GetImports(); err != nil {
		return err
	}
	if _, err = r.GetCodeStats(); err != nil {
		return err
	}
	if _, err = r.GetLintMessages(DefaultLinters); err != nil {
		return err
	}
	if _, err = r.GetTestResults(); err != nil {
		return err
	}
	if _, err = r.GetScore(); err != nil {
		return err
	}
	if _, err = r.GetMetadata(); err != nil {
		return err
	}
	if _, err = r.GetLastUpdate(); err != nil {
		return err
	}
	if _, err = r.GetExecutionTime(); err != nil {
		return err
	}
	return err
}

// ClearCache removes the repository from database.
func (r *Repository) ClearCache() (err error) {
	prefix := fmt.Sprintf("%s-%s", r.Name, r.Branch)
	return r.db.DeleteAllMatchingPrefix([]byte(prefix))
}

// AsMap generates a map out of repository fields.
func (r *Repository) AsMap() map[string]interface{} {
	return map[string]interface{}{
		model.ImportsName:       r.Imports,
		model.CodeStatsName:     r.CodeStats,
		model.LintMessagesName:  r.LintMessages,
		model.TestResultsName:   r.TestResults,
		model.ScoreName:         r.Score,
		model.MetadataName:      r.Metadata,
		model.LastUpdateName:    r.LastUpdate,
		model.ExecutionTimeName: r.ExecutionTime,
	}
}
