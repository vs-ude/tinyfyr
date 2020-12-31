package types

// StackOrder ...
type StackOrder struct {
	Name   string
	Higher []*StackOrder
	Equal  []*StackOrder
}
