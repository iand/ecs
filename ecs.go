package ecs

import (
	"errors"
	"sort"
	"sync"
)

// Entities have components
// Archetypes are combinations of components
// Entities belong to a single archetype (and by inference to all archetypes that consist of a subset of that archetype's components)
type (
	ArchetypeID     struct{ uint64 }
	ComponentTypeID struct{ uint64 }
	EntityID        struct{ uint64 }
)

var (
	// UndefinedArchetypeID is an identifier that can be used as a placeholder for an unknown archetype. It will not
	// match a defined archetype.
	UndefinedArchetypeID = ArchetypeID{0}

	// UndefinedComponentTypeID is an identifier that can be used as a placeholder for an unknown component type. It will not
	// match a defined component type.
	UndefinedComponentTypeID = ComponentTypeID{0}

	// UndefinedEntityID is an identifier that can be used as a placeholder for an unknown entity. It will not
	// match a defined entity.
	UndefinedEntityID = EntityID{0}
)

// NullArchetypeID is the identifier of the special "null" archetype that includes all known entities. This is
// different to the UndefinedArchetypeID.
var NullArchetypeID = GetArchetypeID()

var (
	ErrUnknownEntity        = errors.New("unknown entity")
	ErrUnknownArchetype     = errors.New("unknown archetype")
	ErrMissingComponent     = errors.New("missing component")
	ErrUnknownComponentType = errors.New("unknown component type")
)

// GetArchetypeID returns the ID of the archetype that represents the supplied components.
func GetArchetypeID(cs ...ComponentTypeID) ArchetypeID {
	const (
		// prime64 is the factor in the FNV-1a hash algorithm
		prime64 = 1099511628211
	)

	sort.SliceStable(cs, func(i, j int) bool { return cs[i].uint64 > cs[j].uint64 })

	s := uint64(14695981039346656037)
	for _, c := range cs {
		s ^= uint64(c.uint64 & 0xff)
		s *= prime64

		s ^= uint64((c.uint64 >> 8) & 0xff)
		s *= prime64

		s ^= uint64((c.uint64 >> 16) & 0xff)
		s *= prime64

		s ^= uint64((c.uint64 >> 24) & 0xff)
		s *= prime64

		s ^= uint64((c.uint64 >> 32) & 0xff)
		s *= prime64

		s ^= uint64((c.uint64 >> 40) & 0xff)
		s *= prime64

		s ^= uint64((c.uint64 >> 48) & 0xff)
		s *= prime64

		s ^= uint64(c.uint64 >> 56)
		s *= prime64
	}
	return ArchetypeID{s}
}

func NewComponentTypeID(id uint64) ComponentTypeID {
	return ComponentTypeID{id}
}

type archetype struct {
	muEntities sync.RWMutex
	entities   map[EntityID]struct{}
	components []ComponentTypeID
}

func newArchetype(components []ComponentTypeID) *archetype {
	return &archetype{
		components: components,
		entities:   make(map[EntityID]struct{}),
	}
}

func (a *archetype) AddEntity(eid EntityID) {
	if eid == UndefinedEntityID {
		return
	}
	a.muEntities.Lock()
	defer a.muEntities.Unlock()
	a.entities[eid] = struct{}{}
}

func (a *archetype) RemoveEntity(eid EntityID) {
	if eid == UndefinedEntityID {
		return
	}
	a.muEntities.Lock()
	defer a.muEntities.Unlock()
	delete(a.entities, eid)
}

func (a *archetype) EachEntity(fn func(EntityID) error, opts ...EachOption) error {
	a.muEntities.RLock()
	defer a.muEntities.RUnlock()

	for eid := range a.entities {
		if err := fn(eid); err != nil {
			return err
		}
	}
	return nil
}

func (a *archetype) MatchComponents(comps map[ComponentTypeID]int) bool {
	for _, ct := range a.components {
		if _, ok := comps[ct]; !ok {
			return false
		}
	}

	return true
}

type EntityIndex struct {
	muEntities sync.RWMutex
	entities   map[EntityID]einfo
	elast      uint64

	muArchetypes       sync.RWMutex
	archetypeIndex     map[ArchetypeID]*archetype
	componentTypeIndex map[ComponentTypeID]map[ArchetypeID]struct{}
}

type einfo struct {
	components map[ComponentTypeID]int
	archetypes map[ArchetypeID]struct{}
}

func NewEntityIndex() *EntityIndex {
	return &EntityIndex{
		entities:           make(map[EntityID]einfo),
		archetypeIndex:     make(map[ArchetypeID]*archetype),
		componentTypeIndex: make(map[ComponentTypeID]map[ArchetypeID]struct{}),
	}
}

func (ei *EntityIndex) NewEntity() EntityID {
	ei.muEntities.Lock()
	defer ei.muEntities.Unlock()
	ei.elast++
	eid := EntityID{ei.elast}
	ei.entities[eid] = einfo{
		components: make(map[ComponentTypeID]int),
		archetypes: make(map[ArchetypeID]struct{}),
	}
	return eid
}

func (ei *EntityIndex) DeleteEntity(eid EntityID) (map[ComponentTypeID]int, error) {
	if eid == UndefinedEntityID {
		return nil, ErrUnknownEntity
	}
	ei.muEntities.Lock()
	defer ei.muEntities.Unlock()

	info, ok := ei.entities[eid]
	if !ok {
		return nil, ErrUnknownEntity
	}
	delete(ei.entities, eid)

	for aid := range info.archetypes {
		a, ok := ei.archetypeIndex[aid]
		if ok {
			a.RemoveEntity(eid)
		}
	}

	return info.components, nil
}

func (ei *EntityIndex) GetComponentIndex(eid EntityID, ct ComponentTypeID) (int, bool, error) {
	if eid == UndefinedEntityID {
		return -1, false, ErrUnknownEntity
	}
	ei.muEntities.RLock()
	defer ei.muEntities.RUnlock()
	info, ok := ei.entities[eid]
	if !ok {
		return -1, false, ErrUnknownEntity
	}

	idx, ok := info.components[ct]
	return idx, ok, nil
}

func (ei *EntityIndex) GetComponentIndices(eid EntityID, cts []ComponentTypeID) ([]int, error) {
	if eid == UndefinedEntityID {
		return nil, ErrUnknownEntity
	}
	ei.muEntities.RLock()
	defer ei.muEntities.RUnlock()
	info, ok := ei.entities[eid]
	if !ok {
		return nil, ErrUnknownEntity
	}

	indices := make([]int, len(cts))
	for i, ct := range cts {
		idx, ok := info.components[ct]
		if !ok {
			return nil, ErrMissingComponent
		}
		indices[i] = idx
	}
	return indices, nil
}

func (ei *EntityIndex) SetComponentIndex(eid EntityID, ct ComponentTypeID, idx int) error {
	if eid == UndefinedEntityID {
		return ErrUnknownEntity
	}
	ei.muEntities.Lock()
	defer ei.muEntities.Unlock()
	info, ok := ei.entities[eid]
	if !ok {
		return ErrUnknownEntity
	}

	info.components[ct] = idx
	ei.entities[eid] = info

	ei.muArchetypes.Lock()
	defer ei.muArchetypes.Unlock()
	if aids, ok := ei.componentTypeIndex[ct]; ok {
		for aid := range aids {
			a, ok := ei.archetypeIndex[aid]
			if !ok {
				return ErrUnknownArchetype
			}
			if a.MatchComponents(info.components) {
				a.AddEntity(eid)
				info.archetypes[aid] = struct{}{}
			}
		}
	}

	return nil
}

func (ei *EntityIndex) ClearComponentIndex(eid EntityID, ct ComponentTypeID) (int, bool, error) {
	if eid == UndefinedEntityID {
		return -1, false, ErrUnknownEntity
	}
	ei.muEntities.Lock()
	defer ei.muEntities.Unlock()
	info, ok := ei.entities[eid]
	if !ok {
		return -1, false, ErrUnknownEntity
	}
	idx, ok := info.components[ct]
	if !ok {
		return -1, false, nil
	}
	delete(info.components, ct)

	ei.muArchetypes.Lock()
	defer ei.muArchetypes.Unlock()
	if aids, ok := ei.componentTypeIndex[ct]; ok {
		for aid := range aids {
			a, ok := ei.archetypeIndex[aid]
			if ok {
				a.RemoveEntity(eid)
			}

			delete(info.archetypes, aid)
		}
	}

	ei.entities[eid] = info
	return idx, true, nil
}

func (ei *EntityIndex) EachEntity(aid ArchetypeID, fn func(EntityID) error, opts ...EachOption) error {
	if aid == UndefinedArchetypeID {
		return ErrUnknownArchetype
	}
	if aid == NullArchetypeID {
		ei.muEntities.Lock()
		defer ei.muEntities.Unlock()
		for eid := range ei.entities {
			if err := fn(eid); err != nil {
				return err
			}
		}
		return nil
	}

	ei.muArchetypes.RLock()
	a, ok := ei.archetypeIndex[aid]
	if !ok {
		ei.muArchetypes.RUnlock()
		return ErrUnknownArchetype
	}
	ei.muArchetypes.RUnlock()

	return a.EachEntity(fn)
}

func (ei *EntityIndex) DefineArchetype(components ...ComponentTypeID) (ArchetypeID, error) {
	aid := GetArchetypeID(components...)

	ei.muArchetypes.Lock()
	defer ei.muArchetypes.Unlock()

	ei.archetypeIndex[aid] = newArchetype(components)

	for _, ct := range components {
		aids, ok := ei.componentTypeIndex[ct]
		if !ok {
			aids = make(map[ArchetypeID]struct{})
		}
		aids[aid] = struct{}{}
		ei.componentTypeIndex[ct] = aids
	}

	return aid, nil
}

func (ei *EntityIndex) DeleteArchetype(aid ArchetypeID) error {
	if aid == UndefinedArchetypeID {
		return ErrUnknownArchetype
	}
	ei.muArchetypes.Lock()
	defer ei.muArchetypes.Unlock()

	a, ok := ei.archetypeIndex[aid]
	if !ok {
		return ErrUnknownArchetype
	}

	for _, ct := range a.components {
		aids, ok := ei.componentTypeIndex[ct]
		if !ok {
			continue
		}
		delete(aids, aid)
		ei.componentTypeIndex[ct] = aids
	}

	delete(ei.archetypeIndex, aid)
	return nil
}

type EachOption interface {
	apply(*eachcfg) error
}

type eachcfg struct{}
