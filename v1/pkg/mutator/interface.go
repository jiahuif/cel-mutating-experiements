package mutator

import (
	"github.com/google/cel-go/common/types/ref"
)

type Interface interface {
	ref.Val

	// Parent returns the parent of this mutator, or nil for the root mutator.
	Parent() Interface

	// Identifier returns the identifier that can find this mutator from its parent,
	// or nil for the root mutator
	Identifier() any

	// Merge performs a simple JSON merge from the list that the mutator holds
	// with the given patch. Returns whether the list has been changed, or any
	// error.
	Merge(patch any) ref.Val

	// Remove removes the referring list from its parent.
	// Returns null, or an error.
	Remove() ref.Val
}

type Container interface {
	Interface

	// RemoveChild removes a child that is identified by the given identifier.
	RemoveChild(identifier any) error

	// Child gets the child by the identifier if the child presents.
	Child(identifier any) (any, bool)

	// SetChild replaces the child by the identifier.
	SetChild(identifier any, value any) error
}
