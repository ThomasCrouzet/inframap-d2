package render

import "github.com/ThomasCrouzet/inframap-d2/internal/model"

// Theme defines colors for different server types and elements.
type Theme struct {
	Name   string
	Colors map[string]ThemeColor
}

// ThemeColor defines fill and stroke colors for an element type.
type ThemeColor struct {
	Fill   string
	Stroke string
	Font   string
}

var themes = map[string]*Theme{
	"default": {
		Name: "default",
		Colors: map[string]ThemeColor{
			"production": {Fill: "#FEE2E2", Stroke: "#DC2626", Font: "#991B1B"},
			"lab":        {Fill: "#DCFCE7", Stroke: "#16A34A", Font: "#166534"},
			"local":      {Fill: "#FEF9C3", Stroke: "#CA8A04", Font: "#854D0E"},
			"cluster":    {Fill: "#E0F2FE", Stroke: "#0284C7", Font: "#075985"},
			"hypervisor": {Fill: "#FFF7ED", Stroke: "#EA580C", Font: "#9A3412"},
			"devices":    {Fill: "#F3F4F6", Stroke: "#6B7280", Font: "#374151"},
			"cloud":      {Fill: "#DBEAFE", Stroke: "#2563EB", Font: "#1E40AF"},
			"database":   {Fill: "#EDE9FE", Stroke: "#7C3AED", Font: "#5B21B6"},
			"system":     {Fill: "#E0E7FF", Stroke: "#4F46E5", Font: "#3730A3"},
		},
	},
	"dark": {
		Name: "dark",
		Colors: map[string]ThemeColor{
			"production": {Fill: "#450A0A", Stroke: "#EF4444", Font: "#FCA5A5"},
			"lab":        {Fill: "#052E16", Stroke: "#22C55E", Font: "#86EFAC"},
			"local":      {Fill: "#422006", Stroke: "#EAB308", Font: "#FDE047"},
			"cluster":    {Fill: "#082F49", Stroke: "#0EA5E9", Font: "#7DD3FC"},
			"hypervisor": {Fill: "#431407", Stroke: "#F97316", Font: "#FDBA74"},
			"devices":    {Fill: "#1F2937", Stroke: "#9CA3AF", Font: "#D1D5DB"},
			"cloud":      {Fill: "#1E3A5F", Stroke: "#3B82F6", Font: "#93C5FD"},
			"database":   {Fill: "#2E1065", Stroke: "#A78BFA", Font: "#C4B5FD"},
			"system":     {Fill: "#1E1B4B", Stroke: "#818CF8", Font: "#A5B4FC"},
		},
	},
	"monochrome": {
		Name: "monochrome",
		Colors: map[string]ThemeColor{
			"production": {Fill: "#E5E7EB", Stroke: "#374151", Font: "#111827"},
			"lab":        {Fill: "#F3F4F6", Stroke: "#6B7280", Font: "#374151"},
			"local":      {Fill: "#F9FAFB", Stroke: "#9CA3AF", Font: "#4B5563"},
			"cluster":    {Fill: "#E5E7EB", Stroke: "#4B5563", Font: "#1F2937"},
			"hypervisor": {Fill: "#D1D5DB", Stroke: "#374151", Font: "#111827"},
			"devices":    {Fill: "#F3F4F6", Stroke: "#9CA3AF", Font: "#6B7280"},
			"cloud":      {Fill: "#E5E7EB", Stroke: "#6B7280", Font: "#374151"},
			"database":   {Fill: "#D1D5DB", Stroke: "#4B5563", Font: "#1F2937"},
			"system":     {Fill: "#E5E7EB", Stroke: "#6B7280", Font: "#374151"},
		},
	},
	"ocean": {
		Name: "ocean",
		Colors: map[string]ThemeColor{
			"production": {Fill: "#FEE2E2", Stroke: "#DC2626", Font: "#991B1B"},
			"lab":        {Fill: "#CFFAFE", Stroke: "#0891B2", Font: "#155E75"},
			"local":      {Fill: "#E0F2FE", Stroke: "#0284C7", Font: "#075985"},
			"cluster":    {Fill: "#DBEAFE", Stroke: "#2563EB", Font: "#1E40AF"},
			"hypervisor": {Fill: "#C7D2FE", Stroke: "#4F46E5", Font: "#3730A3"},
			"devices":    {Fill: "#F0F9FF", Stroke: "#38BDF8", Font: "#0369A1"},
			"cloud":      {Fill: "#E0F2FE", Stroke: "#0EA5E9", Font: "#0C4A6E"},
			"database":   {Fill: "#C7D2FE", Stroke: "#6366F1", Font: "#3730A3"},
			"system":     {Fill: "#DBEAFE", Stroke: "#3B82F6", Font: "#1E40AF"},
		},
	},
}

// ThemeNames returns all available theme names.
func ThemeNames() []string {
	names := make([]string, 0, len(themes))
	for name := range themes {
		names = append(names, name)
	}
	return names
}

// GetTheme returns the named theme or the default.
func GetTheme(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["default"]
}

// ColorForServerType returns the theme color for a server type.
func (t *Theme) ColorForServerType(st model.ServerType) ThemeColor {
	if c, ok := t.Colors[string(st)]; ok {
		return c
	}
	return t.Colors["lab"]
}

// ColorForElement returns the theme color for a named element.
func (t *Theme) ColorForElement(name string) ThemeColor {
	if c, ok := t.Colors[name]; ok {
		return c
	}
	return ThemeColor{Fill: "#F9FAFB", Stroke: "#D1D5DB", Font: "#111827"}
}
