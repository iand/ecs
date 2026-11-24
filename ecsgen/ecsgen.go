package ecsgen

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed em.tmpl
var tmpl string

type Data struct {
	Package     string
	ManagerType string
	Components  []Component
	Archetypes  []Archetype
}

type Component struct {
	ID         uint64
	TypeName   string
	TypeIDName string
	OnRemove   string
}

type Archetype struct {
	TypeName   string
	TypeIDName string
	Components []Component
}

type Generator struct {
	data  Data
	fname string

	components []string
	archetypes map[string][]string
}

func (g *Generator) Package(name string) {
	g.data.Package = name
}

func (g *Generator) ManagerName(name string) {
	g.data.ManagerType = name
}

func (g *Generator) Filename(name string) {
	g.fname = name
}

func (g *Generator) AddComponent(typeName string, opts ...ComponentOption) {
	g.components = append(g.components, typeName)
	c := Component{
		ID:         uint64(len(g.components)),
		TypeName:   typeName,
		TypeIDName: fmt.Sprintf("%sComponentTypeID", typeName),
	}

	for _, opt := range opts {
		opt(&c)
	}

	g.data.Components = append(g.data.Components, c)
}

func (g *Generator) AddArchetype(typeName string, componentTypes ...string) {
	if g.archetypes == nil {
		g.archetypes = make(map[string][]string)
	}

	g.archetypes[typeName] = componentTypes

	arch := Archetype{
		TypeName:   typeName,
		TypeIDName: fmt.Sprintf("%sArchetypeID", typeName),
	}
	for _, c := range componentTypes {
		arch.Components = append(arch.Components, Component{
			TypeName:   c,
			TypeIDName: fmt.Sprintf("%sComponentTypeID", c),
		})
	}

	g.data.Archetypes = append(g.data.Archetypes, arch)
}

func (g *Generator) Generate() error {
	if err := g.Validate(); err != nil {
		return err
	}

	t, err := template.New("em").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse embedded template: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}

	fname := filepath.Join(cwd, g.fname)
	fmt.Printf("Writing to %s\n", fname)

	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("could not create output file: %w", err)
	}
	defer f.Close()

	return t.Execute(f, &g.data)
}

func (g *Generator) Validate() error {
	componentTypeIDsSeen := map[uint64]bool{}
	for _, c := range g.data.Components {
		if c.TypeName == "" {
			return fmt.Errorf("missing name for component with id: %d", c.ID)
		}
		if c.ID == 0 {
			return fmt.Errorf("cannot use zero as a component type id : %s", c.TypeName)
		}

		if componentTypeIDsSeen[c.ID] {
			return fmt.Errorf("duplicate component type id: %d (%s)", c.ID, c.TypeName)
		}
		componentTypeIDsSeen[c.ID] = true
	}

	for _, a := range g.data.Archetypes {
		if a.TypeName == "" {
			return fmt.Errorf("missing name for archetype")
		}
		if len(a.Components) < 1 {
			return fmt.Errorf("archetype must at least one component: %s", a.TypeName)
		}
	}

	return nil
}

type ComponentOption func(*Component)

func OnRemove(funcname string) ComponentOption { return func(c *Component) { c.OnRemove = funcname } }

func WithID(id uint64) ComponentOption { return func(c *Component) { c.ID = id } }
