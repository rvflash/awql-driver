package awql

import (
	"database/sql/driver"
	"encoding/csv"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	apiUrl     = "https://adwords.google.com/api/adwords/reportdownload/"
	apiFmt     = "CSV"
	apiTimeout = time.Duration(30 * time.Second)
)

// AwqlStmt is a prepared statement.
type AwqlStmt struct {
	conn  *AwqlConn
	query string
}

// Close closes the statement.
func (s *AwqlStmt) Close() error {
	return nil
}

// NumInput returns the number of placeholder parameters.
func (s *AwqlStmt) NumInput() int {
	return strings.Count(s.query, "?")
}

// Query sends request to Google Adwords API and retrieves its content.
func (s *AwqlStmt) Query(args []driver.Value) (driver.Rows, error) {
	// Binds all the args on the query
	if err := s.bind(args); err != nil {
		return nil, err
	}
	// Saves response in a file named with the hash64 of the query.
	f, err := s.filepath()
	if err != nil {
		return nil, err
	}
	// Downloads the report
	if err := s.download(f); err != nil {
		return nil, err
	}
	// Parse the CSV report.
	d, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	r := csv.NewReader(d)
	rs, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if l := (len(rs) - 1); l > 0 {
		return &AwqlRows{size: uint(l), data: rs[1:], pos: 1}, nil
	}
	return &AwqlRows{}, nil
}

// Exec executes a query that doesn't return rows, such as an INSERT or UPDATE.
func (s *AwqlStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, driver.ErrSkip
}

// bind applies the required argument replacements on the query.
func (s *AwqlStmt) bind(args []driver.Value) error {
	if len(args) != s.NumInput() {
		return ErrQueryBinding
	}
	q := s.query
	for _, v := range args {
		q = strings.Replace(q, "?", fmt.Sprintf("%q", v), 1)
	}
	s.query = q

	return nil
}

// download calls Adwords API and saves response in a file.
func (s *AwqlStmt) download(name string) error {
	rq, err := http.NewRequest(
		"POST", apiUrl+s.conn.opts.version,
		strings.NewReader(url.Values{"__rdquery": {s.query}, "__fmt": {apiFmt}}.Encode()),
	)
	if err != nil {
		return err
	}
	s.conn.client.Timeout = apiTimeout
	rq.Header.Add("Content-Type", "application/x-www-form-urlencoded; param=value")
	rq.Header.Add("Accept", "*/*")
	rq.Header.Add("clientCustomerId", s.conn.adwordsID)
	rq.Header.Add("developerToken", s.conn.developerToken)
	rq.Header.Add("includeZeroImpressions", strconv.FormatBool(s.conn.opts.includeZeroImpressions))
	rq.Header.Add("skipColumnHeader", strconv.FormatBool(s.conn.opts.skipColumnHeader))
	rq.Header.Add("skipReportHeader", strconv.FormatBool(s.conn.opts.skipReportHeader))
	rq.Header.Add("skipReportSummary", strconv.FormatBool(s.conn.opts.skipReportSummary))
	rq.Header.Add("useRawEnumValues", strconv.FormatBool(s.conn.opts.useRawEnumValues))

	// Uses access token to fetch report
	if s.conn.WithAuth() {
		if err := s.conn.RefreshAuth(); err != nil {
			return ErrBadToken
		}
		rq.Header.Add("Authorization", s.conn.oAuth.String())
	}

	// Downloads the report
	resp, err := s.conn.client.Do(rq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Manages response in error
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 0:
			return ErrNoNetwork
		case http.StatusBadRequest:
			out, _ := ioutil.ReadAll(resp.Body)
			return NewApiError(out)
		default:
			return ErrBadNetwork
		}
	}

	// Saves response in a file
	out, err := os.Create(name)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

// filepath returns the filepath to save the response of the query.
func (s *AwqlStmt) filepath() (string, error) {
	h := fnv.New64()
	if _, err := h.Write([]byte(s.query)); err != nil {
		return "", err
	}
	return filepath.Join(
		os.TempDir(), "awql", strconv.FormatUint(h.Sum64(), 10), ".", strings.ToLower(apiFmt),
	), nil
}
