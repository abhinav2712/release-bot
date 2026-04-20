package models

// ReleaseItem represents a single developer's branch entry in the current release.
type ReleaseItem struct {
	DeveloperID   string
	DeveloperName string
	Branch        string
	Title         string
	Status        string
	PRLink        string
	Blocker       string
}

// CurrentRelease holds all state for the active release.
type CurrentRelease struct {
	ThreadID     string
	SummaryMsgID string
	ReleaseType  string
	ReleaseDate  string
	ReleaseNotes string
	Items        []ReleaseItem
}
