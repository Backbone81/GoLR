package backend

// Transition is a single transition on a character range to the next state.
type Transition struct {
	// ByteRange describes the bytes on which to use this transition
	ByteRange ByteRange `json:"byteRange" yaml:"byteRange"`

	// StateIdx is the target state to transition to.
	StateIdx int `json:"stateIdx" yaml:"stateIdx"`
}
