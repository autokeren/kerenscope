package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type Result struct {
	OK     bool           `json:"ok"`
	Data   any            `json:"data,omitempty"`
	Error  string         `json:"error,omitempty"`
}

type Definition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type Tool interface {
	Definition() Definition
	Run(ctx context.Context, args map[string]any) Result
}

type Registry struct {
	tools map[string]Tool
	order []string
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(t Tool) *Registry {
	def := t.Definition()
	r.tools[def.Name] = t
	r.order = append(r.order, def.Name)
	return r
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) Names() []string {
	return append([]string(nil), r.order...)
}

func (r *Registry) Definitions() []Definition {
	out := make([]Definition, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.tools[name].Definition())
	}
	return out
}

func (r *Registry) Catalog() string {
	data, _ := json.MarshalIndent(r.Definitions(), "", " ")
	return string(data)
}

func (r *Registry) Run(ctx context.Context, name string, args map[string]any) Result {
	t, ok := r.Get(name)
	if !ok {
		return Result{OK: false, Error: "tool not found: " + name}
	}
	return t.Run(ctx, args)
}

func argString(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		switch t := v.(type) {
		case string:
			return t
		case float64:
			return trimFloat(t)
		case json.Number:
			return t.String()
		}
	}
	return ""
}

func argInt(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
		if s, ok := v.(string); ok {
			var n int
			if _, err := fmt.Sscan(s, &n); err == nil {
				return n
			}
		}
	}
	return def
}

func argFloat(args map[string]any, key string, def float64) float64 {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return def
}

func trimFloat(f float64) string {
	if f == float64(int64(f)) {
		return jsonInt(f)
	}
	b, _ := json.Marshal(f)
	return string(b)
}

func jsonInt(f float64) string {
	b, _ := json.Marshal(int64(f))
	return string(b)
}
