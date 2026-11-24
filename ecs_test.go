package ecs

import (
	"testing"
)

func TestEntityIndex(t *testing.T) {
	componentType1 := ComponentTypeID{1}
	componentType2 := ComponentTypeID{2}

	collectEntityIDs := func(aid ArchetypeID, ei *EntityIndex) []EntityID {
		eids := make([]EntityID, 0)
		ei.EachEntity(aid, func(eid EntityID) error {
			eids = append(eids, eid)
			return nil
		})
		return eids
	}

	t.Run("new_entity_added_to_null_archetype", func(t *testing.T) {
		ei := NewEntityIndex()
		eid := ei.NewEntity()
		eids := collectEntityIDs(NullArchetypeID, ei)

		if len(eids) != 1 {
			t.Fatalf("got %d entities, wanted 1", len(eids))
		}

		if eids[0] != eid {
			t.Fatalf("got %v entity, wanted %v", eids[0], eid)
		}
	})

	t.Run("entity_with_component_not_removed_from_null_archetype", func(t *testing.T) {
		ei := NewEntityIndex()
		eid := ei.NewEntity()
		ei.SetComponentIndex(eid, componentType1, 1)
		eids := collectEntityIDs(NullArchetypeID, ei)

		if len(eids) != 1 {
			t.Fatalf("got %d entities, wanted 1", len(eids))
		}
	})

	t.Run("entity_last_component_removed_kept_in_null_archetype", func(t *testing.T) {
		ei := NewEntityIndex()
		eid := ei.NewEntity()
		ei.SetComponentIndex(eid, componentType1, 1)
		ei.ClearComponentIndex(eid, componentType1)
		eids := collectEntityIDs(NullArchetypeID, ei)

		if len(eids) != 1 {
			t.Fatalf("got %d entities, wanted 1", len(eids))
		}
	})

	t.Run("new_entity_not_added_to_archetype", func(t *testing.T) {
		ei := NewEntityIndex()
		aid, err := ei.DefineArchetype(componentType1, componentType2)
		if err != nil {
			t.Fatalf("unexpected error defining archetype: %v", err)
		}

		_ = ei.NewEntity()
		eids := collectEntityIDs(aid, ei)

		if len(eids) != 0 {
			t.Fatalf("got %d entities, wanted 0", len(eids))
		}
	})

	t.Run("entity_with_one_component_not_added_to_archetype", func(t *testing.T) {
		ei := NewEntityIndex()
		aid, err := ei.DefineArchetype(componentType1, componentType2)
		if err != nil {
			t.Fatalf("unexpected error defining archetype: %v", err)
		}

		eid := ei.NewEntity()
		ei.SetComponentIndex(eid, componentType1, 1)
		eids := collectEntityIDs(aid, ei)

		if len(eids) != 0 {
			t.Fatalf("got %d entities, wanted 0", len(eids))
		}
	})

	t.Run("entity_with_two_components_added_to_archetype", func(t *testing.T) {
		ei := NewEntityIndex()
		aid, err := ei.DefineArchetype(componentType1, componentType2)
		if err != nil {
			t.Fatalf("unexpected error defining archetype: %v", err)
		}

		eid := ei.NewEntity()
		ei.SetComponentIndex(eid, componentType1, 1)
		ei.SetComponentIndex(eid, componentType2, 4)
		eids := collectEntityIDs(aid, ei)

		if len(eids) != 1 {
			t.Fatalf("got %d entities, wanted 1", len(eids))
		}
	})

	t.Run("entity_with_component_removed_removes_from_archetype", func(t *testing.T) {
		ei := NewEntityIndex()
		aid, err := ei.DefineArchetype(componentType1, componentType2)
		if err != nil {
			t.Fatalf("unexpected error defining archetype: %v", err)
		}

		eid := ei.NewEntity()
		ei.SetComponentIndex(eid, componentType1, 1)
		ei.SetComponentIndex(eid, componentType2, 4)

		ei.ClearComponentIndex(eid, componentType1)
		eids := collectEntityIDs(aid, ei)

		if len(eids) != 0 {
			t.Fatalf("got %d entities, wanted 0", len(eids))
		}
	})

	t.Run("entity_deleted_removes_from_archetype", func(t *testing.T) {
		ei := NewEntityIndex()
		aid, err := ei.DefineArchetype(componentType1, componentType2)
		if err != nil {
			t.Fatalf("unexpected error defining archetype: %v", err)
		}

		eid := ei.NewEntity()
		ei.SetComponentIndex(eid, componentType1, 1)
		ei.SetComponentIndex(eid, componentType2, 4)

		ei.DeleteEntity(eid)
		eids := collectEntityIDs(aid, ei)

		if len(eids) != 0 {
			t.Fatalf("got %d entities, wanted 0", len(eids))
		}
	})
}
