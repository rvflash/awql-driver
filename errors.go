package awql

import (
	"encoding/xml"
	"errors"
)

var (
	ErrIntNoData    = errors.New("InternalError.MISSING_DATA_SOURCE")
	ErrQueryBinding = errors.New("QueryError.BINDING_NOT_MATCH")
	ErrNoNetwork    = errors.New("ConnectionError.NOT_FOUND")
	ErrBadNetwork   = errors.New("ConnectionError.SERVICE_UNAVAILABLE")
	ErrBadToken     = errors.New("ConnectionError.INVALID_ACCESS_TOKEN")
	ErrAdwordsID    = errors.New("ConnectionError.ADWORDS_ID")
	ErrDevToken     = errors.New("ConnectionError.DEVELOPER_TOKEN")
)

// In case of error, Google Adwords API provides more information in a XML response
// @example
// <?xml version="1.0" encoding="UTF-8" standalone="yes"?>
// <reportDownloadError>
// 	<ApiError>
// 		<type>ReportDefinitionError.CUSTOMER_SERVING_TYPE_REPORT_MISMATCH</type>
// 		<trigger></trigger>
// 		<fieldPath>selector</fieldPath>
// 	</ApiError>
// </reportDownloadError>
//
// ApiError represents a Google Report Download Error.
// Voluntary ignores trigger field.
type ApiError struct {
	Type    string `xml:"ApiError>type"`
	Trigger string `xml:"ApiError>trigger"`
	Field   string `xml:"ApiError>fieldPath"`
}

// NewApiError parses a XML document that represents a download report error.
// It returns the given message as error.
func NewApiError(d []byte) error {
	if len(d) == 0 {
		return ErrIntNoData
	}
	e := ApiError{}
	err := xml.Unmarshal(d, &e)
	if err != nil {
		return err
	}
	return errors.New(e.String())
}

// String returns a representation of the api error.
// It implements Stringer interface
func (e *ApiError) String() string {
	switch e.Field {
	case "":
		if e.Trigger == "" {
			return e.Type
		}
		return e.Type + " (" + e.Trigger + ")"
	case "selector":
		return e.Type
	default:
		return e.Type + " on " + e.Field
	}
}
