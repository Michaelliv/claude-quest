package main

// ChestType indicates whether this is a level-up or bonus chest
type ChestType int

const (
	ChestTypeLevelUp ChestType = iota
	ChestTypeBonus
)

// ChestState represents the current state in the chest ceremony
type ChestState int

const (
	ChestClosed ChestState = iota
	ChestWobble
	ChestOpening
	ChestRevealing
	ChestChoosing // Level-up only - player picks from items
	ChestClaiming
	ChestDone
)

// TreasureChest manages the treasure chest opening ceremony
type TreasureChest struct {
	Type        ChestType
	State       ChestState
	Items       []Item  // Items to display (3 for level-up, 1 for bonus)
	SelectedIdx int     // Currently highlighted item
	ClaimedItem *Item   // The item that was claimed
	Timer       float32 // Animation timer for current state
	Reason      string  // For bonus chests, why it was triggered
}

// State durations
const (
	chestWobbleDuration    = 0.8
	chestOpeningDuration   = 0.5
	chestRevealingDuration = 0.6
	chestClaimingDuration  = 1.0
)

// NewLevelUpChest creates a chest for level-up rewards
func NewLevelUpChest(profile *CareerProfile) *TreasureChest {
	items := profile.GetRandomChoices(3)
	return &TreasureChest{
		Type:        ChestTypeLevelUp,
		State:       ChestClosed,
		Items:       items,
		SelectedIdx: 0,
	}
}

// NewBonusChest creates a bonus treasure chest with one random item
func NewBonusChest(profile *CareerProfile, reason string) *TreasureChest {
	pool := profile.GetChoicePool()
	var items []Item
	if len(pool) > 0 {
		// Pick one random item
		items = profile.GetRandomChoices(1)
	}
	return &TreasureChest{
		Type:        ChestTypeBonus,
		State:       ChestClosed,
		Items:       items,
		SelectedIdx: 0,
		Reason:      reason,
	}
}

// Update advances the chest animation state machine
func (c *TreasureChest) Update(dt float32) {
	c.Timer += dt

	switch c.State {
	case ChestClosed:
		// Wait for player to press a key (handled externally)
		// Auto-advance after a brief pause
		if c.Timer > 0.5 {
			c.State = ChestWobble
			c.Timer = 0
		}

	case ChestWobble:
		if c.Timer > chestWobbleDuration {
			c.State = ChestOpening
			c.Timer = 0
		}

	case ChestOpening:
		if c.Timer > chestOpeningDuration {
			c.State = ChestRevealing
			c.Timer = 0
		}

	case ChestRevealing:
		if c.Timer > chestRevealingDuration {
			if c.Type == ChestTypeLevelUp && len(c.Items) > 1 {
				c.State = ChestChoosing
			} else {
				// Bonus chests auto-claim their single item
				c.State = ChestClaiming
				if len(c.Items) > 0 {
					c.ClaimedItem = &c.Items[0]
				}
			}
			c.Timer = 0
		}

	case ChestChoosing:
		// Player selects with arrow keys, confirms with enter
		// Handled externally via SelectItem/ConfirmSelection

	case ChestClaiming:
		if c.Timer > chestClaimingDuration {
			c.State = ChestDone
			c.Timer = 0
		}

	case ChestDone:
		// Chest ceremony complete
	}
}

// SelectNext moves selection to the next item
func (c *TreasureChest) SelectNext() {
	if c.State != ChestChoosing || len(c.Items) == 0 {
		return
	}
	c.SelectedIdx = (c.SelectedIdx + 1) % len(c.Items)
}

// SelectPrev moves selection to the previous item
func (c *TreasureChest) SelectPrev() {
	if c.State != ChestChoosing || len(c.Items) == 0 {
		return
	}
	c.SelectedIdx--
	if c.SelectedIdx < 0 {
		c.SelectedIdx = len(c.Items) - 1
	}
}

// ConfirmSelection claims the currently selected item
func (c *TreasureChest) ConfirmSelection() {
	if c.State != ChestChoosing || len(c.Items) == 0 {
		return
	}
	c.ClaimedItem = &c.Items[c.SelectedIdx]
	c.State = ChestClaiming
	c.Timer = 0
}

// SkipToReveal allows skipping the initial animation
func (c *TreasureChest) SkipToReveal() {
	if c.State == ChestClosed || c.State == ChestWobble {
		c.State = ChestOpening
		c.Timer = 0
	}
}

// IsDone returns true if the chest ceremony is complete
func (c *TreasureChest) IsDone() bool {
	return c.State == ChestDone
}

// IsInteractive returns true if the chest is waiting for player input
func (c *TreasureChest) IsInteractive() bool {
	return c.State == ChestChoosing
}

// GetWobbleOffset returns a wobble animation offset for the chest sprite
func (c *TreasureChest) GetWobbleOffset() float32 {
	if c.State != ChestWobble {
		return 0
	}
	// Oscillating wobble that increases in intensity
	intensity := c.Timer / chestWobbleDuration
	return float32(simpleSinF(float64(c.Timer*20))) * intensity * 3
}

// GetOpenProgress returns 0-1 progress through the opening animation
func (c *TreasureChest) GetOpenProgress() float32 {
	if c.State == ChestOpening {
		return c.Timer / chestOpeningDuration
	}
	if c.State > ChestOpening {
		return 1.0
	}
	return 0
}

// GetRevealProgress returns 0-1 progress through item reveal animation
func (c *TreasureChest) GetRevealProgress() float32 {
	if c.State == ChestRevealing {
		return c.Timer / chestRevealingDuration
	}
	if c.State > ChestRevealing {
		return 1.0
	}
	return 0
}

// GetClaimProgress returns 0-1 progress through claiming animation
func (c *TreasureChest) GetClaimProgress() float32 {
	if c.State == ChestClaiming {
		return c.Timer / chestClaimingDuration
	}
	if c.State > ChestClaiming {
		return 1.0
	}
	return 0
}

// HasItems returns true if there are items to show
func (c *TreasureChest) HasItems() bool {
	return len(c.Items) > 0
}
