package main

import "testing"

// newPickerTestRenderer builds a minimal Renderer suitable for exercising
// syncItemIndices without touching raylib resources. The texture slices are
// left empty — only the *Names slices are inspected by the function under
// test.
func newPickerTestRenderer(profile *CareerProfile, hatNames, faceNames, auraNames, trailNames []string) *Renderer {
	return &Renderer{
		profile:      profile,
		hatNames:     hatNames,
		faceNames:    faceNames,
		auraNames:    auraNames,
		trailNames:   trailNames,
		currentHat:   -1,
		currentFace:  -1,
		currentAura:  -1,
		currentTrail: -1,
	}
}

// loadedHatNamesMissingOne returns the full hat-id list from ItemRegistry
// with one id removed, simulating a packaged asset bundle whose .png for
// that id is missing on disk so loadHats never populated it into hatNames.
func loadedHatNamesMissingOne(missingID string) []string {
	var names []string
	for _, item := range ItemRegistry {
		if item.Slot != SlotHat {
			continue
		}
		if item.ID == missingID {
			continue
		}
		names = append(names, item.ID)
	}
	return names
}

// allOwnedProfile returns a profile that owns every item in ItemRegistry.
func allOwnedProfile() *CareerProfile {
	p := &CareerProfile{OwnedItems: map[string]bool{}}
	for _, item := range ItemRegistry {
		p.OwnedItems[item.ID] = true
	}
	return p
}

// findItemIDAfter returns the id of the first hat in ItemRegistry that
// appears AFTER the supplied id. Useful for picking a known "shifted"
// target in drift tests.
func findItemIDAfter(after string) string {
	seen := false
	for _, item := range ItemRegistry {
		if item.Slot != SlotHat {
			continue
		}
		if seen {
			return item.ID
		}
		if item.ID == after {
			seen = true
		}
	}
	return ""
}

// pickerPositionInLoadedList computes the expected pickerItemIndex value for
// the supplied hat id, mirroring exactly how cycleSlotItem/getSlotItemInfo
// build their items list: 0 = "none", then owned-and-loaded items in
// ItemRegistry order. syncItemIndices must agree with this counting.
func pickerPositionInLoadedList(id string, names []string, profile *CareerProfile) int {
	loaded := map[string]struct{}{}
	for _, n := range names {
		loaded[n] = struct{}{}
	}
	idx := 1
	for _, item := range ItemRegistry {
		if item.Slot != SlotHat {
			continue
		}
		if !profile.IsOwned(item.ID) {
			continue
		}
		if _, ok := loaded[item.ID]; !ok {
			continue
		}
		if item.ID == id {
			return idx
		}
		idx++
	}
	return 0
}

// indexOf returns the position of id in names, or -1.
func indexOf(names []string, id string) int {
	for i, n := range names {
		if n == id {
			return i
		}
	}
	return -1
}

func TestSyncItemIndices_HappyPathAllLoaded(t *testing.T) {
	// No drift: every owned hat is also loaded. pickerItemIndex must match
	// the position in cycleSlotItem's items list (none + loaded-owned).
	profile := allOwnedProfile()
	hatNames := loadedHatNamesMissingOne("") // remove nothing — all loaded
	r := newPickerTestRenderer(profile, hatNames, nil, nil, nil)

	// Pick a mid-list hat and the last hat; both must yield matching positions.
	cases := []string{hatNames[0], hatNames[len(hatNames)/2], hatNames[len(hatNames)-1]}
	for _, id := range cases {
		r.currentHat = indexOf(hatNames, id)
		r.syncItemIndices()
		want := pickerPositionInLoadedList(id, hatNames, profile)
		if got := r.pickerItemIndex[0]; got != want {
			t.Errorf("hat %q: pickerItemIndex[0] = %d; want %d (happy path, no drift)", id, got, want)
		}
	}
}

func TestSyncItemIndices_DriftBetweenRegistryAndLoadedNames(t *testing.T) {
	// Simulate an asset/registry drift where one hat is in ItemRegistry and
	// owned by the profile but absent from hatNames (e.g. its .png wasn't
	// shipped with this build). Every hat in ItemRegistry that comes AFTER
	// the missing one should still resolve to its correct visible position,
	// and the LAST registry hat must not land out of bounds.
	const dropped = "zeus" // hat that exists in ItemRegistry mid-list

	profile := allOwnedProfile()
	hatNames := loadedHatNamesMissingOne(dropped)
	r := newPickerTestRenderer(profile, hatNames, nil, nil, nil)

	// Sanity: hat directly after the dropped one must NOT share its position
	// with anything else. The bug was off-by-one for this hat, and out-of-
	// bounds for the final hat.
	afterDropped := findItemIDAfter(dropped)
	if afterDropped == "" {
		t.Fatalf("test fixture: no hat appears after %q in ItemRegistry", dropped)
	}

	cases := []string{afterDropped, hatNames[len(hatNames)-1]}
	for _, id := range cases {
		r.currentHat = indexOf(hatNames, id)
		if r.currentHat < 0 {
			t.Fatalf("test fixture: %q not present in loaded hatNames", id)
		}
		r.syncItemIndices()
		want := pickerPositionInLoadedList(id, hatNames, profile)
		if got := r.pickerItemIndex[0]; got != want {
			t.Errorf("hat %q: pickerItemIndex[0] = %d; want %d (drift case — %q missing from hatNames)",
				id, got, want, dropped)
		}
	}
}

func TestSyncItemIndices_UnownedItemsNotCounted(t *testing.T) {
	// Items the profile does not own must never advance the position counter,
	// regardless of whether they would be loaded.
	profile := &CareerProfile{OwnedItems: map[string]bool{}}
	hatNames := loadedHatNamesMissingOne("")

	// Own only the last hat in ItemRegistry. Pickers expect position 1 (just
	// after "none") because it is the sole owned-loaded entry.
	var lastHat string
	for _, item := range ItemRegistry {
		if item.Slot == SlotHat {
			lastHat = item.ID
		}
	}
	profile.OwnedItems[lastHat] = true

	r := newPickerTestRenderer(profile, hatNames, nil, nil, nil)
	r.currentHat = indexOf(hatNames, lastHat)

	r.syncItemIndices()
	if got := r.pickerItemIndex[0]; got != 1 {
		t.Errorf("solitary owned hat at end of registry: pickerItemIndex[0] = %d; want 1", got)
	}
}

func TestSyncItemIndices_NoCurrentItemResetsToNone(t *testing.T) {
	// currentHat = -1 (no hat equipped) must always land on position 0.
	profile := allOwnedProfile()
	hatNames := loadedHatNamesMissingOne("")
	r := newPickerTestRenderer(profile, hatNames, nil, nil, nil)
	r.pickerItemIndex[0] = 5 // stale value — must be overwritten

	r.syncItemIndices()
	if got := r.pickerItemIndex[0]; got != 0 {
		t.Errorf("currentHat=-1: pickerItemIndex[0] = %d; want 0", got)
	}
}
