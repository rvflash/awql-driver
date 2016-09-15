package awql

import (
	"database/sql/driver"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	apiUrl = "https://adwords.google.com/api/adwords/reportdownload/"
	apiFmt = "CSV"
)

// AwqlStmt is a prepared statement.
type AwqlStmt struct {
	Conn  *AwqlConn
	Query string
}

// Close closes the statement.
func (s *AwqlStmt) Close() error {
	return nil
}

// NumInput returns the number of placeholder parameters.
func (s *AwqlStmt) NumInput() int {
	return strings.Count(s.Query, "?")
}

// Query sends request to Google Adwords API and retrieves its content.
func (s *AwqlStmt) Query(args []driver.Value) (driver.Rows, error) {
	// Binds all the args on the query
	if err := s.bind(args); err != nil {
		return nil, err
	}
	// Downloads the report and saves response in a file named with the hash64 of the query.
	h := fnv.New64()
	if _, err := h.Write([]byte(s.Query)); err != nil {
		return nil, err
	}
	f := filepath.Join(os.TempDir(), "awql"+h.Sum64()+"."+strings.ToLower(apiFmt))
	if err := s.download(f); err != nil {
		return nil, err
	}
	return nil, nil
}

// Exec executes a query that doesn't return rows, such as an INSERT or UPDATE.
func (s *AwqlStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, driver.ErrSkip
}

// download calls Adwords API and saves response in a file.
func (s *AwqlStmt) download(name string) error {
	rq, err := http.NewRequest(
		"POST", apiUrl+s.Conn.Opts.version,
		strings.NewReader(url.Values{"__rdquery": {s.Query}, "__fmt": {apiFmt}}.Encode()),
	)
	if err != nil {
		return err
	}
	rq.Header.Add("Content-Type", "application/x-www-form-urlencoded; param=value")
	rq.Header.Add("Accept", "*/*")
	rq.Header.Add("developerToken", s.Conn.DeveloperToken)
	rq.Header.Add("clientCustomerId", s.Conn.AdwordsID)
	rq.Header.Add("skipReportHeader", strconv.FormatBool(s.Conn.Opts.skipReportHeader))
	rq.Header.Add("skipColumnHeader", strconv.FormatBool(s.Conn.Opts.skipColumnHeader))
	rq.Header.Add("skipReportSummary", strconv.FormatBool(s.Conn.Opts.skipReportSummary))
	rq.Header.Add("includeZeroImpressions", strconv.FormatBool(s.Conn.Opts.includeZeroImpressions))
	rq.Header.Add("useRawEnumValues", strconv.FormatBool(s.Conn.Opts.useRawEnumValues))

	// Downloads the report
	resp, err := s.Conn.Client.Do(rq)
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
			return NewApiError(&resp.Body)
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

// bind applies the required argument replacements on the query.
func (s *AwqlStmt) bind(args []driver.Value) error {
	if len(args) != s.NumInput() {
		return ErrQueryBinding
	}
	q := s.Query
	for _, v := range args {
		q = strings.Replace(q, "?", fmt.Sprintf("%q", v), 1)
	}
	s.Query = q

	return nil
}
