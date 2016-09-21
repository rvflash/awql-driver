package awql

import (
	"database/sql/driver"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	tokenUrl            = "https:/accounts.google.com/o/oauth2/token"
	tokenTimeout        = time.Duration(4 * time.Second)
	tokenExpiryDelta    = 10 * time.Second
	tokenExpiryDuration = 60 * time.Second
)

// AwqlConn represents a connection to a database and implements driver.Conn.
type AwqlConn struct {
	client         *http.Client
	adwordsID      string
	developerToken string
	oAuth          *AwqlAuth
	opts           *AwqlOpts
}

// Close marks this connection as no longer in use.
func (c *AwqlConn) Close() error {
	c.client = nil
	return nil
}

// Begin is dedicated to start a transaction and awql does not support it.
func (c *AwqlConn) Begin() (driver.Tx, error) {
	return nil, driver.ErrSkip
}

// Prepare returns a prepared statement, bound to this connection.
func (c *AwqlConn) Prepare(q string) (driver.Stmt, error) {
	return &AwqlStmt{conn: c, query: q}, nil
}

// Auth returns an error if it can not download or parse the Google access token.
func (c *AwqlConn) Auth() error {
	if c.oAuth.clientId == "" || c.oAuth.clientSecret == "" || c.oAuth.refreshToken == "" {
		return ErrBadToken
	}
	r, err := c.downloadToken()
	if err != nil {
		return err
	}
	return c.retrieveToken(r)
}

// RefreshAuth returns an error if it can not refresh the access token.
func (c *AwqlConn) RefreshAuth() error {
	if c.oAuth.Valid() {
		return nil
	}
	return c.Auth()
}

// WithAuth returns true if the connection requires an authentification.
func (c *AwqlConn) WithAuth() bool {
	return (c.oAuth.accessToken != "")
}

// NewAuthByToken returns an AwqlAuth struct only based on the access token.
func NewAuthByToken(tk string) (*AwqlAuth, error) {
	if tk == "" {
		return &AwqlAuth{}, io.EOF
	}
	return &AwqlAuth{
		accessToken: tk,
		tokenType:   "Bearer",
		expiry:      time.Now().Add(tokenExpiryDuration),
	}, nil
}

// downloadToken calls Google Auth Api to retrieve an access token.
// @example Google Token
// {
//     "access_token": "ya29.ExaMple",
//     "token_type": "Bearer",
//     "expires_in": 60
// }
func (c *AwqlConn) downloadToken() (io.ReadCloser, error) {
	rq, err := http.NewRequest(
		"POST", tokenUrl,
		strings.NewReader(url.Values{
			"client_id":     {c.oAuth.clientId},
			"client_secret": {c.oAuth.clientSecret},
			"refresh_token": {c.oAuth.refreshToken},
			"grant_type":    {"refresh_token"},
		}.Encode()),
	)
	if err != nil {
		return nil, err
	}
	c.client.Timeout = tokenTimeout
	rq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Retrieves an access token
	resp, err := c.client.Do(rq)
	if err != nil {
		return nil, err
	}

	// Manages response in error
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 0:
			return nil, ErrNoNetwork
		case http.StatusBadRequest:
			return nil, ErrBadToken
		default:
			return nil, ErrBadNetwork
		}
	}
	return resp.Body, nil
}

// retrieveToken parses the JSON response in order to map it to a AwqlToken.
// An error occurs if the JSON is invalid.
func (c *AwqlConn) retrieveToken(d io.ReadCloser) error {
	var tk struct {
		accessToken  string `json:"access_token"`
		expiresInSec int    `json:"expires_in"`
		tokenType    string `json:"token_type"`
	}
	defer d.Close()

	err := json.NewDecoder(d).Decode(&tk)
	if err != nil {
		return ErrBadToken
	}
	if tk.expiresInSec == 0 || tk.accessToken == "" {
		return ErrBadToken
	}
	c.oAuth.accessToken = tk.accessToken
	c.oAuth.tokenType = tk.tokenType
	c.oAuth.expiry = time.Now().Add(time.Duration(tk.expiresInSec) * time.Second)

	return nil
}

// AwqlAuth contains all information to retrieve an access token via OAuth Google.
// It implements Stringer interface
type AwqlAuth struct {
	clientId,
	clientSecret,
	refreshToken,
	accessToken,
	tokenType string
	expiry time.Time
}

// Valid returns in success is the access token is ok.
// The delta in seconds is used to avoid delay expiration of the token.
func (a *AwqlAuth) Valid() bool {
	if a.expiry.IsZero() {
		return false
	}
	return a.expiry.Add(-tokenExpiryDelta).Before(time.Now())
}

// String implements Stringer interface and returns a representation of the access token.
func (a *AwqlAuth) String() string {
	return a.tokenType + " " + a.accessToken
}
