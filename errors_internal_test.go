package awql

import "testing"

var reportDefinitionErrorTests = []struct {
	xml []byte // in
	err string // out
}{
	{[]byte(""), ErrNoDsn.Error()},
	{[]byte(`
	<reportDownloadError>
		<APIError>
			<type>ReportDefinitionError.CUSTOMER_SERVING_TYPE_REPORT_MISMATCH</type>
			<trigger></trigger>
			<fieldPath>selector</fieldPath>
		</APIError>
	</reportDownloadError>`), "ReportDefinitionError.CUSTOMER_SERVING_TYPE_REPORT_MISMATCH"},
	{[]byte(`
	<reportDownloadError>
		<APIError>
			<type>ReportDownloadError.ERROR_GETTING_RESPONSE_FROM_BACKEND</type>
			<trigger>Unable to read report data</trigger>
			<fieldPath></fieldPath>
		</APIError>
	</reportDownloadError>`), "ReportDownloadError.ERROR_GETTING_RESPONSE_FROM_BACKEND (Unable to read report data)"},
	{[]byte(`
	<reportDownloadError>
		<APIError>
			<type>QueryError.DATE_COLUMN_REQUIRES_DURING_CLAUSE</type>
			<trigger></trigger>
			<fieldPath></fieldPath>
		</APIError>
	</reportDownloadError>`), "QueryError.DATE_COLUMN_REQUIRES_DURING_CLAUSE"},
	{[]byte(`
	<reportDownloadError>
		<APIError>
			<type>ReportDefinitionError.INVALID_FIELD_NAME_FOR_REPORT</type>
			<trigger></trigger>
			<fieldPath>CampaignId</fieldPath>
		</APIError>
	</reportDownloadError>`), "ReportDefinitionError.INVALID_FIELD_NAME_FOR_REPORT on CampaignId"},
	{[]byte(`Not a XML`), "EOF"},
}

// TestNewApiError tests the method named NewAPIError.
func TestNewApiError(t *testing.T) {
	for _, e := range reportDefinitionErrorTests {
		if err := NewAPIError(e.xml); err != nil {
			if err.Error() != e.err {
				t.Errorf("Expected the error message %v with %s, received %v", e.err, e.xml, err)
			}
		} else {
			t.Errorf("Expected an error with %v", e.xml)
		}
	}
}
