package awql

// Dsn represents a data source name.
type Dsn struct {
	AdwordsId, ApiVersion,
	DeveloperToken, AccessToken,
	ClientId, ClientSecret,
	RefreshToken string
}

// NewDsn returns a new instance of Dsn.
func NewDsn(id string) *Dsn {
	return &Dsn{AdwordsId: id}
}

// String outputs the data source name as string.
// It implements fmt.Stringer
// @see AdwordsId[:ApiVersion]|DeveloperToken[|AccessToken]
// @see AdwordsId[:ApiVersion]|DeveloperToken[|ClientId][|ClientSecret][|RefreshToken]
func (d *Dsn) String() (n string) {
	if d.AdwordsId == "" {
		return
	}
	n = d.AdwordsId
	if d.ApiVersion != "" {
		n += dsnOptSep + d.ApiVersion
	}
	if d.DeveloperToken != "" {
		n += dsnSep + d.DeveloperToken
	}
	if d.AccessToken != "" {
		n += dsnSep + d.AccessToken
	}
	if d.ClientId != "" {
		n += dsnSep + d.ClientId
	}
	if d.ClientSecret != "" {
		n += dsnSep + d.ClientSecret
	}
	if d.RefreshToken != "" {
		n += dsnSep + d.RefreshToken
	}
	return
}
