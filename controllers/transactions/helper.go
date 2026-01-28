package transactions

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

func round2dp(f float64) float64 {
	return math.Round(f*100) / 100
}

func lowerCamelKey(key string) string {
	if key == "" {
		return key
	}
	r := []rune(key)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func normalizeJSONKeysLowerCamel(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for k, v2 := range t {
			out[lowerCamelKey(k)] = normalizeJSONKeysLowerCamel(v2)
		}
		return out
	case []interface{}:
		for i := range t {
			t[i] = normalizeJSONKeysLowerCamel(t[i])
		}
		return t
	default:
		return v
	}
}

func formatMoney2dp(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch t := v.(type) {
	case float64:
		return round2dp(t)
	case float32:
		return round2dp(float64(t))
	case int:
		return round2dp(float64(t))
	case int64:
		return round2dp(float64(t))
	case int32:
		return round2dp(float64(t))
	case uint:
		return round2dp(float64(t))
	case uint64:
		return round2dp(float64(t))
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return round2dp(f)
		}
		return t.String()
	case []byte:
		return formatMoney2dp(string(t))
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return t
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return round2dp(f)
		}
		return t
	default:
		// Try best-effort string parse for exotic DB types
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "" {
			return v
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return round2dp(f)
		}
		return v
	}
}

func isMoneyKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "amount") ||
		strings.Contains(k, "moneyin") ||
		strings.Contains(k, "moneyout") ||
		strings.Contains(k, "balance") ||
		strings.Contains(k, "fee")
}

// FormatMoneyFields2dp mutates the provided row map, formatting money-related fields
// (amount/moneyIn/moneyOut/balance/fee) to a number (float64) rounded to 2 decimals.
func FormatMoneyFields2dp(row map[string]interface{}) {
	for k, v := range row {
		if !isMoneyKey(k) {
			continue
		}
		row[k] = formatMoney2dp(v)
	}
}
