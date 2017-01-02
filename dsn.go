package awql

// AwqlDsn represents a data source name.
type AwqlDsn struct {
	AdwordsId, ApiVersion,
	DeveloperToken,
	ClientId, ClientSecret,
	RefreshToken string
}

// NewDsn returns a new instance of AwqlDsn.
func NewDsn(id string) *AwqlDsn {
	return &AwqlDsn{AdwordsId: id}
}

// String outputs the data source name as string.
// It implements fmt.Stringer
// @see AdwordsId[:ApiVersion]|DeveloperToken[|ClientId][|ClientSecret][|RefreshToken]
func (d *AwqlDsn) String() (n string) {
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
