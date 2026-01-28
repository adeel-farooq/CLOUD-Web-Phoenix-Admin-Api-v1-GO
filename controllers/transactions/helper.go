package transactions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func formatMoney2dp(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch t := v.(type) {
	case float64:
		return fmt.Sprintf("%.2f", t)
	case float32:
		return fmt.Sprintf("%.2f", float64(t))
	case int:
		return fmt.Sprintf("%.2f", float64(t))
	case int64:
		return fmt.Sprintf("%.2f", float64(t))
	case int32:
		return fmt.Sprintf("%.2f", float64(t))
	case uint:
		return fmt.Sprintf("%.2f", float64(t))
	case uint64:
		return fmt.Sprintf("%.2f", float64(t))
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return fmt.Sprintf("%.2f", f)
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
			return fmt.Sprintf("%.2f", f)
		}
		return t
	default:
		// Try best-effort string parse for exotic DB types
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "" {
			return v
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return fmt.Sprintf("%.2f", f)
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
// (amount/moneyIn/moneyOut/balance/fee) to a string with exactly 2 decimals.
func FormatMoneyFields2dp(row map[string]interface{}) {
	for k, v := range row {
		if !isMoneyKey(k) {
			continue
		}
		row[k] = formatMoney2dp(v)
	}
}
