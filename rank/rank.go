package rank

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/exago/svc/redis"
	redigo "github.com/garyburd/redigo/redis"
)

var (
	errMissingData = errors.New("Not enough data to calculate the rank")

	required = map[string]bool{
		"loc":     true,
		"imports": true,
		"test":    true,
	}
)

// Rank gathers the necessary parts for computing the rank.
type Rank struct {
	conn       redigo.Conn
	repository string

	// Unmarshaled data retrieved from the DB
	loc     map[string]int
	imports []string
	tests   testRunner

	// Output
	Score Score `json:"score"`
}

// New initialises the redis connection.
func New() *Rank {
	return &Rank{
		conn: redis.GetConn(),
	}
}

// SetRepository sets the repository name.
func (rk *Rank) SetRepository(repo string) {
	rk.repository = repo
}

// GetScore returns the project's rank with a few more infos:
// calculated score and details.
func (rk *Rank) GetScore() (interface{}, error) {
	data, err := rk.loadData()
	if err != nil {
		return nil, err
	}

	if err := rk.deserialize(data); err != nil {
		return nil, err
	}

	// Calculate the score
	rk.calcScore()

	// Save the latest score in DB
	if err := rk.save(); err != nil {
		return nil, err
	}

	return rk, err
}

// GetRankFromCache retrieves the rank (A, B, C...) which
func (rk *Rank) GetRankFromCache() (string, error) {
	o, err := rk.conn.Do("HGET", rk.repository, "rank")
	if o == nil {
		return "", errMissingData
	}
	if err != nil {
		return "", err
	}
	b, err := gzipDecode(o.([]byte))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// loadData checks if the data necessary for computing the rank
// is available and if so returns it as a map.
func (rk *Rank) loadData() (map[string][]byte, error) {
	data := map[string][]byte{}
	for idfr := range required {
		o, err := rk.conn.Do("HGET", rk.repository, idfr)
		if o == nil {
			return nil, errMissingData
		}
		if err != nil {
			return nil, err
		}
		b, err := gzipDecode(o.([]byte))
		if err != nil {
			return nil, err
		}
		data[idfr] = b
	}

	return data, nil
}

// save the latest rank in database without TTL.
// The rank is later retrieved for the badge.
func (rk *Rank) save() error {
	_, err := rk.conn.Do("HSET", rk.repository, "rank", rk.Score.Rank)
	return err
}

// deserialize unmarshals the data into the Rank.
func (rk *Rank) deserialize(data map[string][]byte) error {
	if err := json.Unmarshal(stripEnvelope(data["loc"]), &rk.loc); err != nil {
		return err
	}
	if err := json.Unmarshal(stripEnvelope(data["imports"]), &rk.imports); err != nil {
		return err
	}
	if err := json.Unmarshal(stripEnvelope(data["test"]), &rk.tests); err != nil {
		return err
	}
	return nil
}

// stripEnvelope removes the JSend enveloppe to simplify the JSON processing.
func stripEnvelope(data []byte) []byte {
	d := string(data)
	d = strings.Replace(d, `{"data":`, "", 1)
	d = strings.Replace(d, `,"status":"success"}`, "", 1)
	return []byte(d)
}

// gzipDecode decodes the given gzip encoded data.
func gzipDecode(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.Write(data); err != nil {
		return nil, err
	}
	rd, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type testRunner struct {
	Checklist struct {
		Failed []struct {
			Category string `json:"Category"`
			Desc     string `json:"Desc"`
			Name     string `json:"Name"`
		} `json:"Failed"`
		Passed []struct {
			Category string `json:"Category"`
			Desc     string `json:"Desc"`
			Name     string `json:"Name"`
		} `json:"Passed"`
	} `json:"checklist"`
	Packages []struct {
		Coverage      float64 `json:"coverage"`
		ExecutionTime float64 `json:"execution_time"`
		Name          string  `json:"name"`
		Success       bool    `json:"success"`
	} `json:"packages"`
}
