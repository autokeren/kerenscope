package tools

import (
	"fmt"
	"strings"
	"time"
)

type apiRoute struct {
	template    string
	segments     []string
	allowed     []string
	allowedSet   map[string]bool
	hasDateRange bool
}

var apiRoutes []apiRoute

func init() {
	for template, allowed := range sectorsAPIParams {
		route := apiRoute{
			template:    template,
			segments:    splitPath(template),
			allowed:     allowed,
			allowedSet:  map[string]bool{},
			hasDateRange: hasStartEnd(allowed),
		}
		for _, p := range allowed {
			route.allowedSet[p] = true
		}
		apiRoutes = append(apiRoutes, route)
	}
}

func splitPath(p string) []string {
	trimmed := strings.Split(strings.Trim(p, "/"), "/")
	out := make([]string, 0, len(trimmed))
	out = append(out, trimmed...)
	return out
}

func hasStartEnd(allowed []string) bool {
	var hasStart, hasEnd bool
	for _, p := range allowed {
		if p == "start" {
			hasStart = true
		}
		if p == "end" {
			hasEnd = true
		}
	}
	return hasStart && hasEnd
}

func matchRoute(path string) (apiRoute, bool) {
	segments := splitPath(path)
	for _, route := range apiRoutes {
		if len(route.segments) != len(segments) {
			continue
		}
		ok := true
		for i, seg := range route.segments {
			if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
				continue
			}
			if seg != segments[i] {
				ok = false
				break
			}
		}
		if ok {
			return route, true
		}
	}
	return apiRoute{}, false
}

func validateSectorsAPIParams(path string, params map[string]string) (map[string]string, error) {
	route, ok := matchRoute(path)
	if !ok {
		return nil, fmt.Errorf("unknown endpoint %s — check the path against the Sectors v2 catalog in this tool's description", path)
	}
	out := map[string]string{}
	var unknown []string
	for key, value := range params {
		if route.allowedSet[key] {
			out[key] = value
			continue
		}
		if key == "days" && route.hasDateRange {
			out["days-consumed"] = value
			continue
		}
		unknown = append(unknown, key)
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("unsupported parameter(s) %s for %s — allowed: %s", strings.Join(unknown, ", "), route.template, strings.Join(route.allowed, ", "))
	}
	if consumed, ok := out["days-consumed"]; ok {
		delete(out, "days-consumed")
		start, end := dateRangeFromDays(consumed)
		out["start"] = start
		out["end"] = end
	}
	return out, nil
}

func dateRangeFromDays(daysValue string) (string, string) {
	days := 0
	if _, err := fmt.Sscan(daysValue, &days); err != nil || days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -days)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}
