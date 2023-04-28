package ptr

import "golang.org/x/exp/constraints"

// FromType returns pointer for a given input type.
func FromType[T any](in T) *T {
	return &in
}

// Primitives is a constraint that permits any primitive Go types.
type Primitives interface {
	constraints.Complex |
		constraints.Signed |
		constraints.Unsigned |
		constraints.Integer |
		constraints.Float |
		constraints.Ordered | ~bool
}

// ToValue returns value for a given pointer input type.
func ToValue[T Primitives](in *T) T {
	var empty T
	if in == nil {
		return empty
	}
	return *in
}
