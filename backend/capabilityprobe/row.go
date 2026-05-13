package capabilityprobe

import "fmt"

// Row matches capability-probe CSV columns and admin JSON rows.
type Row struct {
	Ts               string `json:"ts"`
	MerchantAPIKeyID int    `json:"merchant_api_key_id"`
	MerchantID       int    `json:"merchant_id"`
	Provider         string `json:"provider"`
	APIFormat        string `json:"api_format"`
	RouteMode        string `json:"route_mode"`
	Probe            string `json:"probe"`
	HTTPCode         string `json:"http_code"`
	OK               string `json:"ok"`
	Note             string `json:"note"`
}

// CSVRecord returns one CSV line (same order as cmd/capability-probe header).
func (r Row) CSVRecord() []string {
	return []string{
		r.Ts,
		fmt.Sprintf("%d", r.MerchantAPIKeyID),
		fmt.Sprintf("%d", r.MerchantID),
		r.Provider,
		r.APIFormat,
		r.RouteMode,
		r.Probe,
		r.HTTPCode,
		r.OK,
		r.Note,
	}
}

func appendRow(rows *[]Row, ts string, keyID, merchantID int, provider, apiFormat, routeMode, probe, httpCode, ok, note string) {
	*rows = append(*rows, Row{
		Ts:               ts,
		MerchantAPIKeyID: keyID,
		MerchantID:       merchantID,
		Provider:         provider,
		APIFormat:        apiFormat,
		RouteMode:        routeMode,
		Probe:            probe,
		HTTPCode:         httpCode,
		OK:               ok,
		Note:             truncate(note, 600),
	})
}

func itoa(v int) string {
	return fmt.Sprintf("%d", v)
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
