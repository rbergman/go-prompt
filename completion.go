package prompt

import (
	"github.com/c-bata/go-prompt/internal/debug"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	shortenSuffix = "..."
	leftPrefix    = " "
	leftSuffix    = " "
	centerPrefix  = " "
	centerSuffix  = " "
	rightPrefix   = " "
	rightSuffix   = " "
)

type SuggestType string

const (
	SuggestTypeDefault = ""
	SuggestTypeLabel   = "label"
)

var (
	leftMargin       = runewidth.StringWidth(leftPrefix + leftSuffix)
	centerMargin     = runewidth.StringWidth(centerPrefix + centerSuffix)
	rightMargin      = runewidth.StringWidth(rightPrefix + rightSuffix)
	completionMargin = leftMargin + centerMargin + rightMargin
)

// Suggest is printed when completing.
type Suggest struct {
	Text                string
	Note                string
	Description         string
	ExpandedDescription string
	Next                string
	Type                SuggestType
	Context             map[string]interface{}  `json:"-"`
	OnSelected          func(s Suggest) Suggest `json:"-"`
	OnCommitted         func(s Suggest) Suggest `json:"-"`
}

func (s Suggest) selected() Suggest {
	if s.OnSelected != nil {
		return s.OnSelected(s)
	}
	return s
}

func (s Suggest) committed() Suggest {
	if s.OnCommitted != nil {
		return s.OnCommitted(s)
	}
	return s
}

func (s Suggest) textWithNext() string {
	if s.Next != "" {
		return s.Text + " " + s.Next
	}
	return s.Text
}

// CompletionManager manages which suggestion is now selected.
type CompletionManager struct {
	selected  int // -1 means nothing one is selected.
	tmp       []Suggest
	max       uint16
	completer Completer

	verticalScroll     int
	wordSeparator      string
	showAtStart        bool
	expandDescriptions bool
}

// GetSelectedSuggestion returns the selected item.
func (c *CompletionManager) GetSelectedSuggestion() (s Suggest, ok bool) {
	if c.selected == -1 {
		return Suggest{}, false
	} else if c.selected < -1 {
		debug.Assert(false, "must not reach here")
		c.selected = -1
		return Suggest{}, false
	} else if len(c.tmp) == 0 {
		return Suggest{}, false
	}
	s = c.tmp[c.selected].selected()
	return s, true
}

// GetSuggestions returns the list of suggestion.
func (c *CompletionManager) GetSuggestions() []Suggest {
	return c.tmp
}

// Reset to select nothing.
func (c *CompletionManager) Reset() {
	c.selected = -1
	c.verticalScroll = 0
	c.Update(*NewDocument())
}

// Update to update the suggestions.
func (c *CompletionManager) Update(in Document) {
	c.tmp = c.completer(in)
}

// Previous to select the previous suggestion item.
func (c *CompletionManager) Previous() {
	if c.verticalScroll == c.selected && c.selected > 0 {
		c.verticalScroll--
	}
	c.selected--
	c.update(c.Previous)
}

// Next to select the next suggestion item.
func (c *CompletionManager) Next() {
	if c.verticalScroll+int(c.max)-1 == c.selected {
		c.verticalScroll++
	}
	c.selected++
	c.update(c.Next)
}

// Completing returns whether the CompletionManager selects something one.
func (c *CompletionManager) Completing() bool {
	return c.selected != -1
}

func (c *CompletionManager) update(skip func()) {
	max := int(c.max)
	if len(c.tmp) < max {
		max = len(c.tmp)
	}

	if c.selected >= len(c.tmp) {
		c.Reset()
	} else if c.selected < -1 {
		c.selected = len(c.tmp) - 1
		c.verticalScroll = len(c.tmp) - max
	}

	if c.selected > -1 && c.tmp[c.selected].Type != SuggestTypeDefault {
		skip()
	}
}

// NewCompletionManager returns initialized CompletionManager object.
func NewCompletionManager(completer Completer, max uint16) *CompletionManager {
	return &CompletionManager{
		selected:  -1,
		max:       max,
		completer: completer,

		verticalScroll: 0,
	}
}
