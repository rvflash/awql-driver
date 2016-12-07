package awql

import (
	"database/sql"
	"database/sql/driver"
	"net/http"
	"strings"
	"time"
)

const apiVersion = "v201609"

// AwqlDriver implements all methods to pretend as a sql database driver.
type AwqlDriver struct{}

// init adds  awql as sql database driver
// @see https://github.com/golang/go/wiki/SQLDrivers
func init() {
	sql.Register("awql", &AwqlDriver{})
}

// Open returns a new connection to the database.
// @see AdwordsId[:ApiVersion]|DeveloperToken[|AccessToken]
// @see AdwordsId[:ApiVersion]|DeveloperToken[|ClientId][|ClientSecret][|RefreshToken]
// @example 123-456-7890:v201607|dEve1op3er7okeN|1234567890-c1i3n7iD.com|c1ien753cr37|1/R3Fr35h-70k3n
func (d *AwqlDriver) Open(dsn string) (driver.Conn, error) {
	conn, err := unmarshal(dsn)
	if err != nil {
		return nil, err
	}
	if conn.oAuth != nil {
		// An authentification is required to connect to Adwords API.
		conn.authenticate()
	}
	return conn, nil
}

// parseDsn returns an pointer to an AwqlConn by parsing a DSN string.
// It throws an error on fails to parse it.
func unmarshal(dsn string) (*AwqlConn, error) {
	var adwordsId = func(s string) string {
		return strings.Split(s, ":")[0]
	}
	var apiVersion = func(s string) string {
		d := strings.Split(s, ":")
		if len(d) == 2 {
			return d[1]
		}
		return ""
	}
	conn := &AwqlConn{}
	if dsn == "" {
		return conn, driver.ErrBadConn
	}

	parts := strings.Split(dsn, "|")
	size := len(parts)
	if size < 2 || size > 5 || size == 4 {
		return conn, driver.ErrBadConn
	}
	// @example 123-456-7890|dEve1op3er7okeN
	conn.client = http.DefaultClient
	conn.adwordsID = adwordsId(parts[0])
	if conn.adwordsID == "" {
		return conn, ErrAdwordsID
	}
	conn.developerToken = parts[1]
	if conn.developerToken == "" {
		return conn, ErrDevToken
	}
	conn.opts = NewOpts(apiVersion(parts[0]))

	var err error
	switch size {
	case 3:
		// @example 123-456-7890|dEve1op3er7okeN|ya29.AcC3s57okeN
		conn.oAuth, err = NewAuthByToken(parts[2])
	case 5:
		// @example 123-456-7890|dEve1op3er7okeN|1234567890-c1i3n7iD.apps.googleusercontent.com|c1ien753cr37|1/R3Fr35h-70k3n
		conn.oAuth, err = NewAuthByClient(parts[2], parts[3], parts[4])
	}
	return conn, err
}

// AwqlToken contains the properties of the Google access token.
type AwqlToken struct {
	AccessToken,
	TokenType string
	Expiry time.Time
}

// AwqlAuthKeys represents the keys used to retrieve an access token.
type AwqlAuthKeys struct {
	ClientId,
	ClientSecret,
	RefreshToken string
}

// AwqlAuth contains all information to deal with an access token via OAuth Google.
// It implements Stringer interface
type AwqlAuth struct {
	AwqlAuthKeys
	AwqlToken
}

// IsSet returns true if the auth struct has keys to refresh access token.
func (a *AwqlAuth) IsSet() bool {
	return a.ClientId != ""
}

// String returns a representation of the access token.
func (a *AwqlAuth) String() string {
	return a.TokenType + " " + a.AccessToken
}

// Valid returns in success is the access token is ok.
// The delta in seconds is used to avoid delay expiration of the token.
func (a *AwqlAuth) Valid() bool {
	if a.Expiry.IsZero() {
		return false
	}
	return a.Expiry.Add(-tokenExpiryDelta).Before(time.Now())
}

// NewAuthByToken returns an AwqlAuth struct only based on the access token.
func NewAuthByToken(tk string) (*AwqlAuth, error) {
	if tk == "" {
		return &AwqlAuth{}, ErrBadToken
	}
	return &AwqlAuth{
		AwqlToken: AwqlToken{
			AccessToken: tk,
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(tokenExpiryDuration),
		},
	}, nil
}

// NewAuthByClient returns an AwqlAuth struct only based on the client keys.
func NewAuthByClient(clientId, clientSecret, refreshToken string) (*AwqlAuth, error) {
	if clientId == "" || clientSecret == "" || refreshToken == "" {
		return &AwqlAuth{}, ErrBadToken
	}
	return &AwqlAuth{
		AwqlAuthKeys: AwqlAuthKeys{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			RefreshToken: refreshToken,
		},
	}, nil
}

// AwqlOpts lists the available Adwords API properties.
type AwqlOpts struct {
	Version string
	SkipReportHeader,
	SkipColumnHeader,
	SkipReportSummary,
	IncludeZeroImpressions,
	UseRawEnumValues bool
}

// NewOpts returns a AwqlOpts with default options.
func NewOpts(version string) *AwqlOpts {
	if version == "" {
		version = apiVersion
	}
	return &AwqlOpts{Version: version, SkipReportHeader: true, SkipReportSummary: true}
}
