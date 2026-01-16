package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// ItemSlot represents different cosmetic slot types
type ItemSlot string

const (
	SlotHat   ItemSlot = "hat"
	SlotFace  ItemSlot = "face"
	SlotAura  ItemSlot = "aura"
	SlotTrail ItemSlot = "trail"
	// Future slots
	// SlotCape  ItemSlot = "cape"
)

// Item represents a cosmetic item that can be unlocked
type Item struct {
	ID       string
	Name     string
	Slot     ItemSlot
	MinLevel int  // Level required to be in choice pool
	Starter  bool // Auto-unlocked at level 1
}

// ItemRegistry contains all unlockable items
var ItemRegistry = []Item{
	// ==================== HATS (15 total) ====================
	{ID: "wizard", Name: "Wizard Hat", Slot: SlotHat, MinLevel: 1, Starter: true},
	{ID: "party", Name: "Party Hat", Slot: SlotHat, MinLevel: 2},
	{ID: "headphones", Name: "Headphones", Slot: SlotHat, MinLevel: 4},
	{ID: "beret", Name: "Artist Beret", Slot: SlotHat, MinLevel: 6},
	{ID: "tophat", Name: "Top Hat", Slot: SlotHat, MinLevel: 8},
	{ID: "zeus", Name: "Zeus Hair", Slot: SlotHat, MinLevel: 10},
	{ID: "catears", Name: "Cat Ears", Slot: SlotHat, MinLevel: 11},
	{ID: "crown", Name: "Royal Crown", Slot: SlotHat, MinLevel: 14},
	{ID: "propeller", Name: "Propeller Hat", Slot: SlotHat, MinLevel: 17},
	{ID: "pirate", Name: "Pirate Hat", Slot: SlotHat, MinLevel: 20},
	{ID: "viking", Name: "Viking Helmet", Slot: SlotHat, MinLevel: 24},
	{ID: "chef", Name: "Chef Toque", Slot: SlotHat, MinLevel: 28},
	{ID: "halo", Name: "Angel Halo", Slot: SlotHat, MinLevel: 32},
	{ID: "jester", Name: "Jester Cap", Slot: SlotHat, MinLevel: 37},
	{ID: "cowboy", Name: "Cowboy Hat", Slot: SlotHat, MinLevel: 42},
	{ID: "fedora", Name: "Fedora", Slot: SlotHat, MinLevel: 48},

	// ==================== FACES (10 total) ====================
	{ID: "mustache", Name: "Mustache", Slot: SlotFace, MinLevel: 1, Starter: true},
	{ID: "dealwithit", Name: "Deal With It", Slot: SlotFace, MinLevel: 3},
	{ID: "monocle", Name: "Monocle", Slot: SlotFace, MinLevel: 7},
	{ID: "pipe", Name: "Sherlock Pipe", Slot: SlotFace, MinLevel: 10},
	{ID: "borat", Name: "Borat Stache", Slot: SlotFace, MinLevel: 13},
	{ID: "eyepatch", Name: "Eye Patch", Slot: SlotFace, MinLevel: 18},
	{ID: "glasses3d", Name: "3D Glasses", Slot: SlotFace, MinLevel: 23},
	{ID: "groucho", Name: "Groucho Glasses", Slot: SlotFace, MinLevel: 30},
	{ID: "bandana", Name: "Ninja Mask", Slot: SlotFace, MinLevel: 38},
	{ID: "wizardbeard", Name: "Wizard Beard", Slot: SlotFace, MinLevel: 45},

	// ==================== AURAS (8 total) ====================
	{ID: "aura_pixel", Name: "Pixel Dust", Slot: SlotAura, MinLevel: 5},
	{ID: "aura_flame", Name: "Flame Aura", Slot: SlotAura, MinLevel: 9},
	{ID: "aura_frost", Name: "Frost Aura", Slot: SlotAura, MinLevel: 15},
	{ID: "aura_electric", Name: "Electric Aura", Slot: SlotAura, MinLevel: 21},
	{ID: "aura_shadow", Name: "Shadow Aura", Slot: SlotAura, MinLevel: 27},
	{ID: "aura_heart", Name: "Heart Aura", Slot: SlotAura, MinLevel: 34},
	{ID: "aura_code", Name: "Matrix Aura", Slot: SlotAura, MinLevel: 40},
	{ID: "aura_rainbow", Name: "Rainbow Aura", Slot: SlotAura, MinLevel: 47},

	// ==================== TRAILS (6 total) ====================
	{ID: "trail_sparkle", Name: "Sparkle Trail", Slot: SlotTrail, MinLevel: 12},
	{ID: "trail_flame", Name: "Flame Trail", Slot: SlotTrail, MinLevel: 19},
	{ID: "trail_frost", Name: "Ice Trail", Slot: SlotTrail, MinLevel: 26},
	{ID: "trail_hearts", Name: "Heart Trail", Slot: SlotTrail, MinLevel: 33},
	{ID: "trail_pixel", Name: "Pixel Trail", Slot: SlotTrail, MinLevel: 41},
	{ID: "trail_rainbow", Name: "Rainbow Trail", Slot: SlotTrail, MinLevel: 50},
}

// CareerProfile stores persistent progression data
type CareerProfile struct {
	// XP & Level
	XP    int `json:"xp"`
	Level int `json:"level"`

	// Item ownership
	OwnedItems    map[string]bool `json:"owned_items"`
	PendingChoice bool            `json:"pending_choice"` // True if level-up choice awaits

	// Lifetime stats
	TotalReads      int            `json:"total_reads"`
	TotalWrites     int            `json:"total_writes"`
	TotalBash       int            `json:"total_bash"`
	BashSuccesses   int            `json:"bash_successes"`
	TotalThinking   map[string]int `json:"total_thinking"` // by level name
	TodosCompleted  int            `json:"todos_completed"`
	AgentsCompleted int            `json:"agents_completed"`
	TokensConsumed  int64          `json:"tokens_consumed"`
	SessionsStarted int            `json:"sessions_started"`

	// Achievements
	PeakFlowCount    int `json:"peak_flow_count"`
	BestBashStreak   int `json:"best_bash_streak"`
	BonusChestsFound int `json:"bonus_chests_found"`

	// Timestamps
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// SessionStats tracks ephemeral per-session data
type SessionStats struct {
	// Activity counts
	Reads          int
	Writes         int
	BashTotal      int
	BashSuccesses  int
	TodosCompleted int
	TotalToolCalls int

	// Flow meter
	FlowMeter       float32 // 0.0 to 1.0
	FlowDecayTimer  float32 // seconds since last event
	FlowPeakReached bool    // True if hit 100% this session

	// Streaks
	CurrentBashStreak int
	BestBashStreak    int

	// Bonus chest
	BonusChestAwarded bool
}

// XP rewards per event type
const (
	XPRead          = 5
	XPWrite         = 10
	XPBashSuccess   = 15
	XPBashFail      = 5
	XPStreakBonus   = 5
	XPThinkNormal   = 10
	XPThinkHard     = 25
	XPThinkBonus    = 10 // per level above normal
	XPTodoComplete  = 20
	XPAgentComplete = 30
	XPFlowPeak      = 100
)

// getProfilePath returns the path to the career profile JSON file
func getProfilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".claude-quest-profile.json"
	}
	return filepath.Join(home, ".claude-quest-profile.json")
}

// LoadProfile loads the career profile from disk, or creates a new one
func LoadProfile() *CareerProfile {
	profile := &CareerProfile{
		OwnedItems:    make(map[string]bool),
		TotalThinking: make(map[string]int),
		FirstSeen:     time.Now(),
		LastSeen:      time.Now(),
	}

	data, err := os.ReadFile(getProfilePath())
	if err != nil {
		// New profile - grant starter items
		profile.grantStarterItems()
		return profile
	}

	if err := json.Unmarshal(data, profile); err != nil {
		// Corrupted file - start fresh
		profile.grantStarterItems()
		return profile
	}

	// Ensure maps are initialized (in case of old profile format)
	if profile.OwnedItems == nil {
		profile.OwnedItems = make(map[string]bool)
	}
	if profile.TotalThinking == nil {
		profile.TotalThinking = make(map[string]int)
	}

	// Ensure starter items are owned (migration for existing profiles)
	profile.grantStarterItems()

	return profile
}

// grantStarterItems ensures all starter items are owned
func (p *CareerProfile) grantStarterItems() {
	for _, item := range ItemRegistry {
		if item.Starter {
			p.OwnedItems[item.ID] = true
		}
	}
}

// Save writes the profile to disk atomically (temp file + rename)
func (p *CareerProfile) Save() error {
	p.LastSeen = time.Now()

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	profilePath := getProfilePath()
	tempPath := profilePath + ".tmp"

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, profilePath)
}

// XPForLevel returns the total XP required to reach a given level
// Uses quadratic curve: level 1 = 100 XP, level 50 = 250,000 XP total
func XPForLevel(level int) int {
	if level <= 0 {
		return 0
	}
	return 100 * level * level
}

// LevelFromXP calculates current level from total XP
func LevelFromXP(xp int) int {
	level := 0
	for XPForLevel(level+1) <= xp {
		level++
	}
	return level
}

// XPToNextLevel returns XP needed to reach the next level
func (p *CareerProfile) XPToNextLevel() int {
	return XPForLevel(p.Level+1) - p.XP
}

// XPProgress returns progress toward next level as 0.0 to 1.0
func (p *CareerProfile) XPProgress() float32 {
	currentLevelXP := XPForLevel(p.Level)
	nextLevelXP := XPForLevel(p.Level + 1)
	if nextLevelXP == currentLevelXP {
		return 1.0
	}
	return float32(p.XP-currentLevelXP) / float32(nextLevelXP-currentLevelXP)
}

// AddXP grants XP and returns true if a level-up occurred
func (p *CareerProfile) AddXP(amount int) bool {
	oldLevel := p.Level
	p.XP += amount
	p.Level = LevelFromXP(p.XP)

	if p.Level > oldLevel {
		p.PendingChoice = true
		return true
	}
	return false
}

// RecordRead tracks a read operation and grants XP
func (p *CareerProfile) RecordRead() bool {
	p.TotalReads++
	return p.AddXP(XPRead)
}

// RecordWrite tracks a write operation and grants XP
func (p *CareerProfile) RecordWrite() bool {
	p.TotalWrites++
	return p.AddXP(XPWrite)
}

// RecordBash tracks a bash operation and grants XP
func (p *CareerProfile) RecordBash(success bool, streak int) bool {
	p.TotalBash++
	xp := XPBashFail
	if success {
		p.BashSuccesses++
		xp = XPBashSuccess
		if streak > 1 {
			xp += XPStreakBonus
		}
		if streak > p.BestBashStreak {
			p.BestBashStreak = streak
		}
	}
	return p.AddXP(xp)
}

// RecordThinking tracks a thinking event and grants XP
func (p *CareerProfile) RecordThinking(level ThinkLevel) bool {
	levelName := "normal"
	xp := XPThinkNormal

	switch level {
	case ThinkHard:
		levelName = "hard"
		xp = XPThinkHard
	case ThinkHarder:
		levelName = "harder"
		xp = XPThinkHard + XPThinkBonus
	case ThinkUltra:
		levelName = "ultra"
		xp = XPThinkHard + XPThinkBonus*2
	}

	p.TotalThinking[levelName]++
	return p.AddXP(xp)
}

// RecordTodoComplete tracks a todo completion and grants XP
func (p *CareerProfile) RecordTodoComplete() bool {
	p.TodosCompleted++
	return p.AddXP(XPTodoComplete)
}

// RecordAgentComplete tracks an agent completion and grants XP
func (p *CareerProfile) RecordAgentComplete() bool {
	p.AgentsCompleted++
	return p.AddXP(XPAgentComplete)
}

// RecordFlowPeak tracks hitting 100% flow and grants XP
func (p *CareerProfile) RecordFlowPeak() bool {
	p.PeakFlowCount++
	return p.AddXP(XPFlowPeak)
}

// RecordTokens tracks token consumption
func (p *CareerProfile) RecordTokens(count int) {
	p.TokensConsumed += int64(count)
}

// GetChoicePool returns items available for level-up choice
func (p *CareerProfile) GetChoicePool() []Item {
	var pool []Item
	for _, item := range ItemRegistry {
		// Item is in pool if: at/below player level AND not yet owned AND not starter
		if !item.Starter && item.MinLevel <= p.Level && !p.OwnedItems[item.ID] {
			pool = append(pool, item)
		}
	}
	return pool
}

// GetRandomChoices returns n random items from the choice pool
func (p *CareerProfile) GetRandomChoices(n int) []Item {
	pool := p.GetChoicePool()
	if len(pool) <= n {
		return pool
	}

	// Fisher-Yates shuffle and take first n
	shuffled := make([]Item, len(pool))
	copy(shuffled, pool)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:n]
}

// ClaimItem adds an item to owned items
func (p *CareerProfile) ClaimItem(itemID string) {
	p.OwnedItems[itemID] = true
	p.PendingChoice = false
}

// IsOwned checks if an item is owned
func (p *CareerProfile) IsOwned(itemID string) bool {
	return p.OwnedItems[itemID]
}

// GetOwnedItems returns all owned items for a given slot
func (p *CareerProfile) GetOwnedItems(slot ItemSlot) []Item {
	var owned []Item
	for _, item := range ItemRegistry {
		if item.Slot == slot && p.OwnedItems[item.ID] {
			owned = append(owned, item)
		}
	}
	return owned
}

// GetLockedItems returns all locked items for a given slot
func (p *CareerProfile) GetLockedItems(slot ItemSlot) []Item {
	var locked []Item
	for _, item := range ItemRegistry {
		if item.Slot == slot && !p.OwnedItems[item.ID] {
			locked = append(locked, item)
		}
	}
	return locked
}

// GetItemByID finds an item by its ID
func GetItemByID(id string) *Item {
	for i := range ItemRegistry {
		if ItemRegistry[i].ID == id {
			return &ItemRegistry[i]
		}
	}
	return nil
}

// BonusChestTrigger defines a condition that can trigger a bonus chest
type BonusChestTrigger struct {
	Name   string
	Check  func(s *SessionStats) bool
	Chance float32
}

// BonusChestTriggers defines all possible bonus chest triggers
var BonusChestTriggers = []BonusChestTrigger{
	{"Flow Peak", func(s *SessionStats) bool { return s.FlowPeakReached }, 0.30},
	{"Bash Streak", func(s *SessionStats) bool { return s.BestBashStreak >= 10 }, 0.20},
	{"Todo Master", func(s *SessionStats) bool { return s.TodosCompleted >= 5 }, 0.25},
	{"Marathon", func(s *SessionStats) bool { return s.TotalToolCalls >= 200 }, 0.40},
}

// CheckBonusChest checks if a bonus chest should be awarded
func (s *SessionStats) CheckBonusChest() (bool, string) {
	if s.BonusChestAwarded {
		return false, ""
	}

	for _, trigger := range BonusChestTriggers {
		if trigger.Check(s) && rand.Float32() < trigger.Chance {
			s.BonusChestAwarded = true
			return true, trigger.Name
		}
	}
	return false, ""
}

// UpdateFlow updates the flow meter based on activity
func (s *SessionStats) UpdateFlow(dt float32, hadActivity bool) bool {
	peakedNow := false

	if hadActivity {
		s.FlowMeter += 0.05
		if s.FlowMeter >= 1.0 {
			s.FlowMeter = 1.0
			if !s.FlowPeakReached {
				s.FlowPeakReached = true
				peakedNow = true
			}
		}
		s.FlowDecayTimer = 0
	} else {
		s.FlowDecayTimer += dt
		if s.FlowDecayTimer > 5.0 {
			s.FlowMeter -= dt * 0.03
			if s.FlowMeter < 0 {
				s.FlowMeter = 0
			}
		}
	}

	return peakedNow
}

// RecordBashResult updates bash streak tracking
func (s *SessionStats) RecordBashResult(success bool) {
	s.BashTotal++
	s.TotalToolCalls++

	if success {
		s.BashSuccesses++
		s.CurrentBashStreak++
		if s.CurrentBashStreak > s.BestBashStreak {
			s.BestBashStreak = s.CurrentBashStreak
		}
	} else {
		s.CurrentBashStreak = 0
	}
}
