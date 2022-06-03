package render

import (
	sqldriver "database/sql/driver"
	"fmt"
	"io"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/siddontang/go/hack"
)

type rows struct {
	*mysql.Resultset

	columns []string
	step    int
}

func newRows(r *mysql.Resultset) (*rows, error) {
	if r == nil {
		return nil, fmt.Errorf("invalid mysql query, no correct result")
	}

	rs := new(rows)
	rs.Resultset = r

	rs.columns = make([]string, len(r.Fields))

	for i, f := range r.Fields {
		rs.columns[i] = hack.String(f.Name)
	}
	rs.step = 0

	return rs, nil
}

func (r *rows) Columns() []string {
	return r.columns
}

func (r *rows) Close() error {
	r.step = -1
	return nil
}

func (r *rows) Next(dest []sqldriver.Value) error {
	if r.step >= r.Resultset.RowNumber() {
		return io.EOF
	} else if r.step == -1 {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < r.Resultset.ColumnNumber(); i++ {
		value, err := r.Resultset.GetValue(r.step, i)
		if err != nil {
			return err
		}

		dest[i] = sqldriver.Value(value)
	}

	r.step++

	return nil
}
