package generator

type UidGenerator interface {
	// GetUID returns a unique ID
	GetUID() (uint64, error)
	// ParseUID returns a string of the unique ID
	ParseUID(uid uint64) string
}
