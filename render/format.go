package render

import "fmt"

// FormatArgs formats arguments into string slice.
// Ref: https://github.com/go-mysql-org/go-mysql/blob/33ea963610607f7b5505fd39d0955b78039ef783/server/stmt.go#L186
func FormatArgs(args []interface{}) []string {
	var results []string
	for _, arg := range args {
		if arg == nil {
			results = append(results, "NULL")
			continue
		}
		switch x := arg.(type) {
		case int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			results = append(results, fmt.Sprintf("%d", x))
		case float32, float64:
			results = append(results, fmt.Sprintf("%f", x))
		case string:
			results = append(results, fmt.Sprintf("%q", x))
		case []byte:
			results = append(results, fmt.Sprintf("%q", string(x)))
		default:
			results = append(results, fmt.Sprintf("%v", x))
		}
	}
	return results
}
