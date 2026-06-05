package scene

// PlaneDoc describes one stack plane for documentation (generated from code).
type PlaneDoc struct {
	Name     string `json:"name"`
	Contents string `json:"contents"`
	Purpose  string `json:"purpose"`
}

// PlaneCatalog returns stack planes in paint order (back → front).
func PlaneCatalog() []PlaneDoc {
	return []PlaneDoc{
		{Name: PlaneBackground.String(), Contents: "Canvas bounds", Purpose: "Whitespace and density rules"},
		{Name: PlaneLane.String(), Contents: "lane, container swimlanes", Purpose: "Grouping backdrop"},
		{Name: PlaneEdge.String(), Contents: "Connector paths", Purpose: "Routing plane checks"},
		{Name: PlaneStructure.String(), Contents: "frame, group", Purpose: "Structural grouping"},
		{Name: PlaneAnnotation.String(), Contents: "infobox, info, tip, warning, note, textbox", Purpose: "Callouts without blocking flow"},
		{Name: PlaneNode.String(), Contents: "Primary flow shapes (box, cloud, …)", Purpose: "Main diagram content"},
		{Name: PlaneLabel.String(), Contents: "Edge label boxes", Purpose: "Horizontal / vertical label placement"},
		{Name: PlaneChrome.String(), Contents: "Title / subtitle band", Purpose: "Visual hierarchy"},
	}
}

// StackModelDescription is the one-line stack model for agents.
func StackModelDescription() string {
	names := make([]string, len(PlaneCatalog()))
	for i, p := range PlaneCatalog() {
		names[i] = p.Name
	}
	out := names[0]
	for i := 1; i < len(names); i++ {
		out += " → " + names[i]
	}
	return "Diagrams are validated as stacked 2D planes: " + out +
		". Collision and routing checks project onto reduced planes."
}
