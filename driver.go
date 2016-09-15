package awql

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"net/http"
	"strings"
)

const apiVersion = "v201607"

// AwqlDriver implements all methods to pretend as a sql database driver.
type AwqlDriver struct{}

// init adds finally awql as sql database driver
// @see https://github.com/golang/go/wiki/SQLDrivers
func init() {
	sql.Register("awql", &AwqlDriver{})
}

// AwqlOpts lists the available Adwords API properties.
type AwqlOpts struct {
	version string
	skipReportHeader,
	skipColumnHeader,
	skipReportSummary,
	includeZeroImpressions,
	useRawEnumValues bool
}

// NewOpts returns a AwqlOpts with default options.
func NewOpts(version string) *AwqlOpts {
	return &AwqlOpts{version: version, skipReportHeader: true, skipReportSummary: true}
}

// AwqlAuth contains all information to retrieve an access token via OAuth Google.
type AwqlAuth struct {
	ClientId,
	ClientSecret,
	RefreshToken string
}

// Open returns a new connection to the database.
// @see AdwordsId[:ApiVersion]|DeveloperToken[|AccessToken]
// @see AdwordsId[:ApiVersion]|DeveloperToken[|ClientId][|ClientSecret][|RefreshToken]
// @example 123-456-7890:v201607|dEve1op3er7okeN|1234567890-c1i3n7iD.com|c1ien753cr37|1/R3Fr35h-70k3n
func (d *AwqlDriver) Open(dsn string) (driver.Conn, error) {
	if dsn == "" {
		return nil, driver.ErrBadConn
	}
	var adwordsId = func(s string) string {
		return strings.Split(s, ":")
	}
	var apiVersion = func(s string) string {
		if v := strings.SplitAfter(s, ":"); v == "" {
			return apiVersion
		} else {
			return v
		}
	}
	parts := strings.Split(dsn, "|")
	switch len(parts) {
	case 2:
		// @example 123-456-7890|dEve1op3er7okeN
		return &AwqlConn{
			Client: http.DefaultClient, AdwordsID: adwordsId(parts[0]),
			DeveloperToken: parts[1], Opts: NewOpts(apiVersion(parts[0])),
		}, nil
	case 3:
		// @example 123-456-7890|dEve1op3er7okeN|ya29.AcC3s57okeN
		return &AwqlConn{
			Client: http.DefaultClient, AdwordsID: adwordsId(parts[0]),
			DeveloperToken: parts[1], OAuth: &AwqlAuth{RefreshToken: parts[2]},
			Opts: NewOpts(apiVersion(parts[0])),
		}, nil
	case 4:
		// @example 123-456-7890|dEve1op3er7okeN|1234567890-c1i3n7iD.apps.googleusercontent.com|c1ien753cr37|1/R3Fr35h-70k3n
		return &AwqlConn{
			Client: http.DefaultClient, AdwordsID: adwordsId(parts[0]),
			DeveloperToken: parts[1], OAuth: &AwqlAuth{ClientId: parts[2], ClientSecret: parts[3]},
			Opts: NewOpts(apiVersion(parts[0])),
		}, nil
	default:
		return nil, driver.ErrBadConn
	}
}

// AwqlConn represents a connection to a database and implements driver.Conn.
type AwqlConn struct {
	Client         *http.Client
	AdwordsID      string
	DeveloperToken string
	OAuth          *AwqlAuth
	Opts           *AwqlOpts
}

// Close marks this connection as no longer in use.
func (c *AwqlConn) Close() error {
	c.Client = nil
	return nil
}

// Begin is dedicated to start a transaction and awql does not support it.
func (c *AwqlConn) Begin() (driver.Tx, error) {
	return nil, driver.ErrSkip
}

// Prepare returns a prepared statement, bound to this connection.
func (c *AwqlConn) Prepare(q string) (driver.Stmt, error) {
	return &AwqlStmt{Conn: c, Query: q}, nil
}
