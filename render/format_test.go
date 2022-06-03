package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatArgs(t *testing.T) {
	args := []interface{}{
		int8(1),
		uint8(2),
		int16(3),
		uint16(4),
		int32(5),
		uint32(6),
		int64(7),
		uint64(8),
		"hello string",
		[]byte("hello bytes"),
	}

	results := FormatArgs(args)
	expected := []string{
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		`"hello string"`,
		`"hello bytes"`,
	}

	assert.Equal(t, expected, results)
}
