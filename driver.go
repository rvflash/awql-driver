package awql

import (
	"database/sql"
	"database/sql/driver"
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

// Open returns a new connection to the database.
// @see AdwordsId[:ApiVersion]|DeveloperToken[|AccessToken]
// @see AdwordsId[:ApiVersion]|DeveloperToken[|ClientId][|ClientSecret][|RefreshToken]
// @example 123-456-7890:v201607|dEve1op3er7okeN|1234567890-c1i3n7iD.com|c1ien753cr37|1/R3Fr35h-70k3n
func (d *AwqlDriver) Open(dsn string) (driver.Conn, error) {
	if dsn == "" {
		return nil, driver.ErrBadConn
	}
	var adwordsId = func(s string) string {
		return strings.Split(s, ":")[0]
	}
	var apiVersion = func(s string) string {
		d := strings.Split(s, ":")
		if len(d) == 2 {
			return d[1]
		}
		return apiVersion
	}
	var err error

	parts := strings.Split(dsn, "|")
	switch len(parts) {
	case 2:
		// @example 123-456-7890|dEve1op3er7okeN
		return &AwqlConn{
			client: http.DefaultClient, adwordsID: adwordsId(parts[0]),
			developerToken: parts[1], opts: NewOpts(apiVersion(parts[0])),
		}, nil
	case 3:
		// @example 123-456-7890|dEve1op3er7okeN|ya29.AcC3s57okeN
		c := &AwqlConn{
			client: http.DefaultClient, adwordsID: adwordsId(parts[0]),
			developerToken: parts[1], opts: NewOpts(apiVersion(parts[0])),
		}
		c.oAuth, err = NewAuthByToken(parts[2])
		if err != nil {
			return nil, ErrBadToken
		}
		return c, nil
	case 5:
		// @example 123-456-7890|dEve1op3er7okeN|1234567890-c1i3n7iD.apps.googleusercontent.com|c1ien753cr37|1/R3Fr35h-70k3n
		c := &AwqlConn{
			client: http.DefaultClient, adwordsID: adwordsId(parts[0]),
			developerToken: parts[1], opts: NewOpts(apiVersion(parts[0])),
			oAuth: &AwqlAuth{clientId: parts[2], clientSecret: parts[3], refreshToken: parts[4]},
		}
		if err = c.Auth(); err != nil {
			return nil, ErrBadToken
		}
		return c, nil
	default:
		return nil, driver.ErrBadConn
	}
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
