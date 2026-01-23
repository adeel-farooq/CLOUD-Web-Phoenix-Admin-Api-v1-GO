package transfer

import (
	"encoding/json"
)

func toInt(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		// best-effort
		var i int
		_ = json.Unmarshal([]byte(x), &i)
		return i
	default:
		return 0
	}
}
