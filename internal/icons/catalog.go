package icons

import "sort"

// Entry documents one icon for render and sceno docs icons.
type Entry struct {
	ID              string   `json:"id"`
	Category        string   `json:"category"`
	Label           string   `json:"label"`
	Use             string   `json:"use"`
	SuggestedShapes []string `json:"suggested_shapes,omitempty"`
	DefaultIconPos  string   `json:"default_icon_pos,omitempty"`
}

// svg path fragments (24×24 viewBox, stroke icons).
var svgPaths = map[string]string{
	"cloud":    `<path d="M17.5 19H9a7 7 0 1 1-.5-14 9 9 0 0 1 9 9 2.5 2.5 0 0 1 0 5Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"database": `<ellipse cx="12" cy="5.5" rx="7" ry="3" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M5 5.5v13c0 1.66 3.13 3 7 3s7-1.34 7-3v-13" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M5 12c0 1.66 3.13 3 7 3s7-1.34 7-3" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"server":   `<rect x="4" y="4" width="16" height="6" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.75"/><rect x="4" y="14" width="16" height="6" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.75"/><circle cx="8" cy="7" r=".75" fill="currentColor"/><circle cx="8" cy="17" r=".75" fill="currentColor"/>`,
	"user":     `<circle cx="12" cy="8" r="3.5" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M5 20c0-3.87 3.13-7 7-7s7 3.13 7 7" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"users":    `<path d="M9 12a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M16 13a3 3 0 1 0 0-6" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M3 20c0-2.76 2.69-5 6-5M13 20c0-2.2 2.24-4 5-4" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"lock":     `<rect x="6" y="11" width="12" height="9" rx="2" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M8 11V8a4 4 0 0 1 8 0v3" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"shield":   `<path d="M12 3 5 6v6c0 4.42 3 7.56 7 8 4-0.44 7-3.58 7-8V6l-7-3Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/>`,
	"workflow": `<circle cx="6" cy="12" r="2.5" fill="none" stroke="currentColor" stroke-width="1.75"/><circle cx="18" cy="6" r="2.5" fill="none" stroke="currentColor" stroke-width="1.75"/><circle cx="18" cy="18" r="2.5" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M8.5 11 15 7M8.5 13 15 17" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"queue":    `<path d="M5 7h14M5 12h14M5 17h10" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"api":      `<path d="M8 8 4 12l4 4M16 8l4 4-4 4M14 5l-4 14" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"storage":  `<path d="M4 7V5a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v2M4 7v10a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M10 11h4" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"k8s":      `<path d="M12 3 4 7v10l8 4 8-4V7l-8-4Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/><circle cx="12" cy="12" r="2" fill="currentColor"/>`,
	"policy":   `<path d="M12 3 5 6v5c0 3.5 3 6 7 7 4-1 7-3.5 7-7V6l-7-3Z" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M9 12l2 2 4-4" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"info":     `<circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M12 11v5M12 8h.01" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"globe":    `<circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"git":      `<circle cx="6" cy="6" r="2" fill="none" stroke="currentColor" stroke-width="1.75"/><circle cx="6" cy="18" r="2" fill="none" stroke="currentColor" stroke-width="1.75"/><circle cx="18" cy="12" r="2" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M6 8v8M8 18h8" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"docker":   `<path d="M4 14h2v2H4zM7 14h2v2H7zM10 14h2v2h-2zM7 11h2v2H7zM10 11h2v2h-2zM13 11h2v2h-2zM10 8h2v2h-2zM13 8h2v2h-2zM16 8h2v2h-2z" fill="currentColor"/><path d="M2 15h15a4 4 0 0 1 4 4v1H6a4 4 0 0 1-4-4v-1z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/>`,
	"settings": `<circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"chart":    `<path d="M4 19V5M4 19h16M8 17V9M12 17V7M16 17v-4" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"terminal": `<path d="M4 6h16v12H4z" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M7 10l3 2-3 2M11 14h5" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"code":     `<path d="M9 8 5 12l4 4M15 8l4 4-4 4" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"folder":   `<path d="M4 8h5l2 2h9v8H4z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/>`,
	"network":  `<rect x="3" y="3" width="7" height="7" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.75"/><rect x="14" y="3" width="7" height="7" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.75"/><rect x="8.5" y="14" width="7" height="7" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M10 6.5h4M6.5 10v7M17.5 10v4M12 14h3.5" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"mail":     `<rect x="4" y="6" width="16" height="12" rx="2" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="m4 8 8 6 8-6" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/>`,
	"key":      `<circle cx="8" cy="15" r="4" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M12 15h8M16 15v3M20 15v2" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"bell":     `<path d="M12 3a5 5 0 0 1 5 5v3l2 3H5l2-3V8a5 5 0 0 1 5-5Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/><path d="M10 20a2 2 0 0 0 4 0" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"check":    `<circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M8 12l2.5 2.5L16 9" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"zap":      `<path d="M13 2 5 14h6l-1 8 8-12h-6l1-8Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/>`,
	"link":     `<path d="M10 14a4 4 0 0 1 0-6l2-2a4 4 0 1 1 6 6l-1 1M14 10a4 4 0 0 1 0 6l-2 2a4 4 0 1 1-6-6l1-1" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"cpu":      `<rect x="5" y="5" width="14" height="14" rx="2" fill="none" stroke="currentColor" stroke-width="1.75"/><rect x="9" y="9" width="6" height="6" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M9 2v3M15 2v3M9 19v3M15 19v3M2 9h3M2 15h3M19 9h3M19 15h3" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"layers":   `<path d="M12 3 3 8l9 5 9-5-9-5Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/><path d="M3 12l9 5 9-5M3 16l9 5 9-5" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/>`,
	"box":      `<path d="M12 3 4 7v10l8 4 8-4V7l-8-4Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/><path d="M12 12 20 7M12 12v9M12 12 4 7" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/>`,
}

// catalog metadata (order preserved for docs).
var catalogMeta = []Entry{
	{ID: "user", Category: "people", Label: "User", Use: "Single person, actor, persona", SuggestedShapes: []string{"actor", "ellipse", "box"}, DefaultIconPos: "top-left"},
	{ID: "users", Category: "people", Label: "Users", Use: "Teams, groups, developers", SuggestedShapes: []string{"actor", "box"}, DefaultIconPos: "top-left"},
	{ID: "api", Category: "integration", Label: "API", Use: "HTTP APIs, gateways, code interfaces", SuggestedShapes: []string{"box", "hexagon"}, DefaultIconPos: "top-left"},
	{ID: "workflow", Category: "integration", Label: "Workflow", Use: "Pipelines, orchestration, portals", SuggestedShapes: []string{"box", "pill"}, DefaultIconPos: "top-left"},
	{ID: "queue", Category: "integration", Label: "Queue", Use: "Job queues, async buffers", SuggestedShapes: []string{"box", "cylinder"}, DefaultIconPos: "top-left"},
	{ID: "link", Category: "integration", Label: "Link", Use: "Integrations, webhooks, coupling", SuggestedShapes: []string{"box", "cloud"}, DefaultIconPos: "top-left"},
	{ID: "server", Category: "compute", Label: "Server", Use: "VMs, runners, bare metal", SuggestedShapes: []string{"box"}, DefaultIconPos: "top-left"},
	{ID: "cpu", Category: "compute", Label: "CPU", Use: "Workers, functions, processing", SuggestedShapes: []string{"box", "hexagon"}, DefaultIconPos: "top-left"},
	{ID: "k8s", Category: "compute", Label: "Kubernetes", Use: "Clusters, EKS/GKE/AKS components", SuggestedShapes: []string{"hexagon", "box"}, DefaultIconPos: "top-left"},
	{ID: "docker", Category: "compute", Label: "Docker", Use: "Containers, images", SuggestedShapes: []string{"box", "hexagon"}, DefaultIconPos: "top-left"},
	{ID: "cloud", Category: "compute", Label: "Cloud", Use: "Managed services, SaaS, regions", SuggestedShapes: []string{"cloud", "hexagon"}, DefaultIconPos: "top-left"},
	{ID: "globe", Category: "compute", Label: "Globe", Use: "Internet, CDN, public endpoints", SuggestedShapes: []string{"cloud", "hexagon"}, DefaultIconPos: "top-left"},
	{ID: "database", Category: "data", Label: "Database", Use: "SQL/NoSQL stores", SuggestedShapes: []string{"cylinder"}, DefaultIconPos: "top-left"},
	{ID: "storage", Category: "data", Label: "Storage", Use: "Buckets, volumes, blob stores", SuggestedShapes: []string{"cylinder", "box"}, DefaultIconPos: "top-left"},
	{ID: "folder", Category: "data", Label: "Folder", Use: "Repos, directories, assets", SuggestedShapes: []string{"document", "box"}, DefaultIconPos: "top-left"},
	{ID: "chart", Category: "data", Label: "Chart", Use: "Metrics, analytics, dashboards", SuggestedShapes: []string{"box", "infobox"}, DefaultIconPos: "top-left"},
	{ID: "lock", Category: "security", Label: "Lock", Use: "Secrets, encryption, auth", SuggestedShapes: []string{"box", "cylinder"}, DefaultIconPos: "top-left"},
	{ID: "shield", Category: "security", Label: "Shield", Use: "Security controls, WAF", SuggestedShapes: []string{"box", "hexagon"}, DefaultIconPos: "top-left"},
	{ID: "policy", Category: "security", Label: "Policy", Use: "Policy-as-code, compliance", SuggestedShapes: []string{"infobox", "box"}, DefaultIconPos: "top-left"},
	{ID: "key", Category: "security", Label: "Key", Use: "API keys, credentials", SuggestedShapes: []string{"box"}, DefaultIconPos: "top-left"},
	{ID: "info", Category: "annotation", Label: "Info", Use: "Callouts, tips (pairs with info/tip shapes)", SuggestedShapes: []string{"info", "infobox", "note"}, DefaultIconPos: "top-left"},
	{ID: "check", Category: "annotation", Label: "Check", Use: "Success, validation passed", SuggestedShapes: []string{"tip", "pill"}, DefaultIconPos: "top-left"},
	{ID: "bell", Category: "annotation", Label: "Bell", Use: "Alerts, notifications", SuggestedShapes: []string{"infobox", "warning"}, DefaultIconPos: "top-left"},
	{ID: "zap", Category: "annotation", Label: "Zap", Use: "Events, triggers, fast path", SuggestedShapes: []string{"pill", "box"}, DefaultIconPos: "top-left"},
	{ID: "git", Category: "dev", Label: "Git", Use: "Repositories, version control", SuggestedShapes: []string{"document", "box"}, DefaultIconPos: "top-left"},
	{ID: "code", Category: "dev", Label: "Code", Use: "Modules, libraries (not API gateway)", SuggestedShapes: []string{"box", "document"}, DefaultIconPos: "top-left"},
	{ID: "terminal", Category: "dev", Label: "Terminal", Use: "CLI, shell, runners", SuggestedShapes: []string{"box", "code"}, DefaultIconPos: "top-left"},
	{ID: "settings", Category: "dev", Label: "Settings", Use: "Config, admin, control plane", SuggestedShapes: []string{"box", "infobox"}, DefaultIconPos: "top-left"},
	{ID: "network", Category: "network", Label: "Network", Use: "VPC, subnets, topology", SuggestedShapes: []string{"lane", "box", "cloud"}, DefaultIconPos: "top-left"},
	{ID: "layers", Category: "network", Label: "Layers", Use: "Stacked infra, planes", SuggestedShapes: []string{"lane", "frame"}, DefaultIconPos: "top-left"},
	{ID: "box", Category: "network", Label: "Package", Use: "Components, artifacts, bundles", SuggestedShapes: []string{"box", "hexagon"}, DefaultIconPos: "top-left"},
	{ID: "mail", Category: "integration", Label: "Mail", Use: "Email, notifications", SuggestedShapes: []string{"box", "cloud"}, DefaultIconPos: "top-left"},
}

func init() {
	paths = svgPaths
}

// Catalog returns documented icons in display order.
func Catalog() []Entry {
	out := make([]Entry, len(catalogMeta))
	copy(out, catalogMeta)
	return out
}

// Categories returns unique category names in catalog order.
func Categories() []string {
	seen := map[string]struct{}{}
	var out []string
	for _, e := range catalogMeta {
		if _, ok := seen[e.Category]; ok {
			continue
		}
		seen[e.Category] = struct{}{}
		out = append(out, e.Category)
	}
	return out
}

// ByCategory groups catalog entries.
func ByCategory() map[string][]Entry {
	m := map[string][]Entry{}
	for _, e := range catalogMeta {
		m[e.Category] = append(m[e.Category], e)
	}
	return m
}

// DocTips returns authoring hints for icons.
func DocTips() []string {
	return []string{
		"Use icon=name on shape lines — never invent names; run sceno docs icons --json",
		"iconPos=top-left (default) for box cards; iconPos=top for narrow columns; iconPos=center on pills",
		"Pair icons with shape kind: database→cylinder, cloud→cloud, policy→infobox",
		"One icon per primary node; avoid icons on every shape in dense diagrams",
		"Icons render at 20px in polished style with label column clearance",
	}
}

// SortedIDs returns all icon ids sorted.
func SortedIDs() []string {
	out := make([]string, 0, len(svgPaths))
	for k := range svgPaths {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
