package ui

import (
	"regexp"
	"strings"
)

var (
	nodeRe   = regexp.MustCompile(`<node\b[^>]*(?:/>|>)`)
	boundsRe = regexp.MustCompile(`\[(\d+),(\d+)\]\[(\d+),(\d+)\]`)
	textAttrRe        = regexp.MustCompile(`text="([^"]*)"`)
	contentDescAttrRe = regexp.MustCompile(`content-desc="([^"]*)"`)
	resourceIDAttrRe  = regexp.MustCompile(`resource-id="([^"]*)"`)
	classAttrRe       = regexp.MustCompile(`class="([^"]*)"`)
	packageAttrRe     = regexp.MustCompile(`package="([^"]*)"`)
	clickableAttrRe   = regexp.MustCompile(`clickable="([^"]*)"`)
	enabledAttrRe     = regexp.MustCompile(`enabled="([^"]*)"`)
	focusedAttrRe     = regexp.MustCompile(`focused="([^"]*)"`)
)

var attrRes = map[string]*regexp.Regexp{
	"text":         textAttrRe,
	"content-desc": contentDescAttrRe,
	"resource-id":  resourceIDAttrRe,
	"class":        classAttrRe,
	"package":      packageAttrRe,
	"clickable":    clickableAttrRe,
	"enabled":      enabledAttrRe,
	"focused":      focusedAttrRe,
}

type Element struct {
	Text        string
	ContentDesc string
	ResourceID  string
	ClassName   string
	Package     string
	Clickable   bool
	Enabled     bool
	Focused     bool
	Password    bool
	Bounds      [4]int
}

func (e Element) Label() string {
	if e.Text != "" {
		return strings.TrimSpace(e.Text)
	}
	return strings.TrimSpace(e.ContentDesc)
}

func (e Element) Center() (int, int) {
	return (e.Bounds[0] + e.Bounds[2]) / 2, (e.Bounds[1] + e.Bounds[3]) / 2
}

func (e Element) Width() int {
	return max(0, e.Bounds[2]-e.Bounds[0])
}

func (e Element) Height() int {
	return max(0, e.Bounds[3]-e.Bounds[1])
}

type Snapshot struct {
	XML      string
	Elements []Element
}

func ParseHierarchy(xml string) Snapshot {
	if xml == "" || !strings.Contains(xml, "node") {
		return Snapshot{}
	}
	var elements []Element
	for _, tag := range nodeRe.FindAllString(xml, -1) {
		bm := boundsRe.FindStringSubmatch(attr(tag, "bounds"))
		if len(bm) != 5 {
			continue
		}
		x1, y1 := atoi(bm[1]), atoi(bm[2])
		x2, y2 := atoi(bm[3]), atoi(bm[4])
		elements = append(elements, Element{
			Text:        attr(tag, "text"),
			ContentDesc: attr(tag, "content-desc"),
			ResourceID:  attr(tag, "resource-id"),
			ClassName:   attr(tag, "class"),
			Package:     attr(tag, "package"),
			Clickable:   attr(tag, "clickable") == "true",
			Enabled:     attr(tag, "enabled") != "false",
			Focused:     attr(tag, "focused") == "true",
			Password:    attr(tag, "password") == "true",
			Bounds:      [4]int{x1, y1, x2, y2},
		})
	}
	return Snapshot{XML: xml, Elements: elements}
}

func attr(tag, name string) string {
	re, ok := attrRes[name]
	if !ok {
		re = regexp.MustCompile(regexp.QuoteMeta(name) + `="([^"]*)"`)
		attrRes[name] = re
	}
	m := re.FindStringSubmatch(tag)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Join(strings.Fields(s), " ")
	return s
}

type FindQuery struct {
	Texts            []string
	ContentDescs     []string
	ResourceIDs      []string
	Prefer           string // first, bottom, top
	PreferClickable  bool
	MinCenterY       int
	MaxCenterY       int
}

type Resolved struct {
	Element Element
	Label   string
	Bounds  [4]int
}

func (r Resolved) Center() (int, int) {
	return r.Element.Center()
}

type Resolver struct {
	Threshold int
}

// DefaultMatchThreshold is the default fuzzy text match score for UI resolution.
const DefaultMatchThreshold = 70

func NewResolver(threshold int) *Resolver {
	if threshold <= 0 {
		threshold = DefaultMatchThreshold
	}
	return &Resolver{Threshold: threshold}
}

// NewDefaultResolver returns a resolver using DefaultMatchThreshold.
func NewDefaultResolver() *Resolver {
	return NewResolver(DefaultMatchThreshold)
}

func (r *Resolver) Find(snap Snapshot, q FindQuery) *Resolved {
	var candidates []Resolved
	for _, elem := range snap.Elements {
		if q.PreferClickable && !elem.Clickable {
			continue
		}
		if !elem.Enabled || elem.Width() <= 0 {
			continue
		}
		_, cy := elem.Center()
		if q.MinCenterY > 0 && cy < q.MinCenterY {
			continue
		}
		if q.MaxCenterY > 0 && cy > q.MaxCenterY {
			continue
		}
		label := elem.Label()
		nlabel := Normalize(label)
		ndesc := Normalize(elem.ContentDesc)
		for _, t := range q.Texts {
			if t == "" {
				continue
			}
			if label == t || nlabel == Normalize(t) || strings.Contains(nlabel, Normalize(t)) {
				candidates = append(candidates, Resolved{Element: elem, Label: label, Bounds: elem.Bounds})
				break
			}
		}
		for _, d := range q.ContentDescs {
			if d == "" {
				continue
			}
			nd := Normalize(d)
			if ndesc == nd || strings.Contains(ndesc, nd) {
				candidates = append(candidates, Resolved{Element: elem, Label: elem.ContentDesc, Bounds: elem.Bounds})
				break
			}
		}
		for _, rid := range q.ResourceIDs {
			if rid != "" && strings.Contains(elem.ResourceID, rid) {
				candidates = append(candidates, Resolved{Element: elem, Label: label, Bounds: elem.Bounds})
			}
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	return selectBest(candidates, q.Prefer)
}

func (r *Resolver) TextExists(snap Snapshot, texts []string) bool {
	q := FindQuery{Texts: texts, PreferClickable: false}
	if r.Find(snap, q) != nil {
		return true
	}
	lower := strings.ToLower(snap.XML)
	for _, t := range texts {
		if t != "" && strings.Contains(lower, strings.ToLower(t)) {
			return true
		}
	}
	return false
}

func selectBest(candidates []Resolved, prefer string) *Resolved {
	if len(candidates) == 1 {
		return &candidates[0]
	}
	best := candidates[0]
	_, bestY := best.Center()
	for _, c := range candidates[1:] {
		_, cy := c.Center()
		switch prefer {
		case "bottom":
			if cy > bestY {
				best, bestY = c, cy
			}
		case "top":
			if cy < bestY {
				best, bestY = c, cy
			}
		case "left":
			_, bx := best.Center()
			cx, _ := c.Center()
			if cx < bx {
				best = c
			}
		case "right":
			_, bx := best.Center()
			cx, _ := c.Center()
			if cx > bx {
				best = c
			}
		default:
			// first match wins
		}
	}
	return &best
}
