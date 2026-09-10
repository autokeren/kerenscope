package tools

import (
	"context"
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/autokeren/kerenscope/internal/sectors"
)

var pathRe = regexp.MustCompile(`^[a-zA-Z0-9/_\-{}\.]+$`)

type SectorsAPITool struct{ Client *sectors.Client }

func (t SectorsAPITool) Definition() Definition {
	return Definition{
		Name: "sectors_api",
		Description: "Generic escape hatch to any Sectors Financial API v2 endpoint, for data not covered by the dedicated tools. " +
			"GET only; the path must be a documented v2 endpoint. Endpoints include: " +
			"/v2/company/corporate-actions/{symbol}/, /v2/company/get-segments/{symbol}/, /v2/company/shareholders-composition/{symbol}/, " +
			"/v2/companies/ (screener: where, q, order_by, limit), /v2/free-float/, /v2/industries/, /v2/subindustries/, " +
			"/v2/most-traded/, /v2/close/, /v2/idx-total/, /v2/listing-performance/{symbol}/, /v2/brokers/, /v2/brokers/top/, " +
			"/v2/broker-activity/{broker_code}/, /v2/broker-activity/{broker_code}/top/, /v2/suspensions/, /v2/companies/quarterly-financial-dates/, " +
			"and the mining extension (/v2/mining/...: companies, sites, licenses, license-auctions, commodities, exports). " +
			"Note: results from this tool are raw evidence only — they are not run through the deterministic metrics engine. " +
			"Prefer the dedicated tools whenever one fits.",
		Parameters: object(map[string]any{
			"path":   strProp("Endpoint path starting with /v2/, e.g. /v2/company/corporate-actions/BBCA/"),
			"params": map[string]any{"type": "object", "description": "Query parameters (values will be stringified)", "additionalProperties": map[string]any{"type": "string"}},
		}, "path"),
	}
}

func (t SectorsAPITool) Run(ctx context.Context, args map[string]any) Result {
	path := strings.TrimSpace(argString(args, "path"))
	if path == "" {
		return Result{OK: false, Error: "path is required"}
	}
	if strings.Contains(path, "://") || strings.Contains(path, "?") {
		return Result{OK: false, Error: "path must be a bare v2 path, not a full URL"}
	}
	path = "/" + strings.TrimLeft(path, "/")
	if !strings.HasPrefix(path, "/v2/") {
		return Result{OK: false, Error: "path must start with /v2/"}
	}
	if !pathRe.MatchString(path) {
		return Result{OK: false, Error: "path contains invalid characters"}
	}
	params := url.Values{}
	if raw, ok := args["params"].(map[string]any); ok {
		for k, v := range raw {
			switch t := v.(type) {
			case string:
				params.Set(k, t)
			case float64:
				params.Set(k, strconv.FormatFloat(t, 'f', -1, 64))
			case bool:
				params.Set(k, strconv.FormatBool(t))
			default:
				data, err := json.Marshal(v)
				if err == nil {
					params.Set(k, string(data))
				}
			}
		}
	}
	var out json.RawMessage
	if err := t.Client.GetJSON(ctx, path, params, ttlForPath(path), &out); err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: out}
}

func ttlForPath(path string) time.Duration {
	switch {
	case strings.Contains(path, "/news") || strings.Contains(path, "/suspensions"):
		return 3 * time.Hour
	case strings.Contains(path, "/daily") || strings.Contains(path, "/foreign-flow") ||
		strings.Contains(path, "/broker") || strings.Contains(path, "/close") || strings.Contains(path, "/most-traded"):
		return 12 * time.Hour
	default:
		return 24 * time.Hour
	}
}

