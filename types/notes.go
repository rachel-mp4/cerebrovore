package types

type Note struct {
	Type        string
	Name        string
	Description string
}

type Patch struct {
	Release   string
	Timestamp string
	Notes     []Note
}
