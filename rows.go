package awql

import (
	"database/sql/driver"
	"io"
)

// AwqlRows is an iterator over an executed query's results.
type AwqlRows struct {
	pos, size uint
	data      [][]string
}

// Close usual closes the rows iterator.
func (r *AwqlRows) Close() error {
	return nil
}

// Columns returns the names of the columns.
func (r *AwqlRows) Columns() []string {
	return r.data[0]
}

// Next is called to populate the next row of data into the provided slice.
func (r *AwqlRows) Next(dest []driver.Value) error {
	if r.pos == r.size {
		return io.EOF
	}
	for k, v := range r.data[r.pos] {
		dest[k] = v
	}
	r.pos++

	return nil
}
