package toot

type Visibility int

const (
	VisibilityPrivate Visibility = iota
	VisibilityUnlisted
	VisibilityPublic
	VisibilityDirect
)

type Status struct {
	Text       string
	MediaIDs   []string
	ReplyToID  string
	Sensitive  bool
	Spoiler    string
	Visibility Visibility
}
