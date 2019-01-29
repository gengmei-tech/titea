package types

// Namespace avoid package cycle include
type Namespace struct {
	// Name group.service
	Name string `json:"name"`

	Type uint8 `json:"type"`
	// Group
	Group string `json:"group"`

	// Service
	Service string `json:"service"`

	// Creator
	Creator string `json:"creator"`

	// CreateAt
	CreateAt int64 `json:"createAt"`

	// DBIndex
	Index uint64 `json:"index"`
}
