package awql_test

import (
	"database/sql/driver"
	"reflect"
	"testing"
	awql "github.com/rvflash/awql-driver"
)

var rowsTests = []struct {
	rows    *awql.AwqlRows
	columns []string
}{
	{&awql.AwqlRows{Size: 2, Data: [][]string{{"id", "name"}, {"19", "rv"}}}, []string{"id", "name"}},
}

// TestAwqlRows_Close tests the method Close on AwqlRows struct.
func TestAwqlRows_Close(t *testing.T) {
	for _, rt := range rowsTests {
		if err := rt.rows.Close(); err != nil {
			t.Errorf("Expected no error when we close the rows, received %v", err)
		}
	}
}

// TestAwqlRows_Columns tests the method Columns on AwqlRows struct.
func TestAwqlRows_Columns(t *testing.T) {
	for _, rt := range rowsTests {
		if c := rt.rows.Columns(); !reflect.DeepEqual(c, rt.columns) {
			t.Errorf("Expected %v as colums, received %v", rt.columns, c)
		}
	}
}

// TestAwqlRows_Next tests the method Next on AwqlRows struct.
func TestAwqlRows_Next(t *testing.T) {
	for _, rs := range rowsTests {
		dest := make([]driver.Value, len(rs.rows.Columns()))
		if err := rs.rows.Next(dest); err != nil {
			t.Errorf("Expected no error when we get the first row, received %v", err)
		} else if dest[0] != rs.columns[0] || dest[1] != rs.columns[1] {
			t.Errorf("Expected %v as colums, received %v, with err %v", rs.columns, dest, err)
		}
	}
}
