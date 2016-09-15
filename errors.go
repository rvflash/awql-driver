package awql

import (
	"encoding/xml"
	"errors"
)

var (
	ErrQueryBinding = errors.New("QueryError.BINDING_NOT_MATCH")
	ErrNoNetwork    = errors.New("ConnectionError.NOT_FOUND")
	ErrBadNetwork   = errors.New("ConnectionError.SERVICE_UNAVAILABLE")
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
	Type  string `xml:"ApiError>type"`
	Field string `xml:"ApiError>fieldPath"`
}

func NewApiError(s *string) error {
	e := ApiError{}
	err := xml.Unmarshal([]byte(s), &e)
	if err != nil {
		return err
	}
	switch e.Field {
	case "":
		fallthrough
	case "selector":
		return errors.New(e.Type)
	default:
		return errors.New(e.Type + " on " + e.Field)
	}
}
