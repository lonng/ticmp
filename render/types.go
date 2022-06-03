package render

import (
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
)

type (
	QueryResult struct {
		Result   *mysql.Result
		Error    error
		Duration time.Duration
	}

	Frame struct {
		Ident string
		Query string
		Args  []interface{}
		TiDB  QueryResult
		MySQL QueryResult
	}

	Render interface {
		Push(frame *Frame)
	}
)
