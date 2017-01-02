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
	tokenExpiryDuration = 60 * time.Minute
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
	// Resets client
	c.client = nil
	return nil
}

// Begin is dedicated to start a transaction and awql does not support it.
func (c *AwqlConn) Begin() (driver.Tx, error) {
	return nil, driver.ErrSkip
}

// Prepare returns a prepared statement, bound to this connection.
func (c *AwqlConn) Prepare(q string) (driver.Stmt, error) {
	if q == "" {
		// No query to prepare.
		return nil, io.EOF
	}
	return &AwqlStmt{conn: c, query: q}, nil
}

// Auth returns an error if it can not download or parse the Google access token.
func (c *AwqlConn) authenticate() error {
	if c.oAuth == nil || c.oAuth.Valid() {
		// Authentification is not required or already validated.
		return nil
	}
	if !c.oAuth.IsSet() {
		// No client information to refresh the token.
		return ErrBadToken
	}
	d, err := c.downloadToken()
	if err != nil {
		return err
	}
	return c.retrieveToken(d)
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
			"client_id":     {c.oAuth.ClientId},
			"client_secret": {c.oAuth.ClientSecret},
			"refresh_token": {c.oAuth.RefreshToken},
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
		// Unable to parse the JSON response.
		return ErrBadToken
	}
	// Invalid format of the token.
	if tk.expiresInSec == 0 || tk.accessToken == "" {
		return ErrBadToken
	}
	c.oAuth.AccessToken = tk.accessToken
	c.oAuth.TokenType = tk.tokenType
	c.oAuth.Expiry = time.Now().Add(time.Duration(tk.expiresInSec) * time.Second)

	return nil
}
