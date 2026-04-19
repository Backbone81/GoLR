package frontend

import "errors"

// NodeKind is a type describing the kind of regular expression node.
type NodeKind int

const (
	NodeAny NodeKind = iota
	NodeCharClass
	NodeConcat
	NodeLiteral
	NodeOneOrMore
	NodeOptional
	NodeOr
	NodeRepetition
	NodeZeroOrMore
)

// Node is a single node of a regular expression.
type Node struct {
	Kind NodeKind `json:"kind" yaml:"kind"`

	Any        Any        `json:"any" yaml:"any"`
	CharClass  CharClass  `json:"charClass" yaml:"charClass"`
	Concat     Concat     `json:"concat" yaml:"concat"`
	Literal    Literal    `json:"literal" yaml:"literal"`
	OneOrMore  OneOrMore  `json:"oneOrMore" yaml:"oneOrMore"`
	Optional   Optional   `json:"optional" yaml:"optional"`
	Or         Or         `json:"or" yaml:"or"`
	Repetition Repetition `json:"repetition" yaml:"repetition"`
	ZeroOrMore ZeroOrMore `json:"zeroOrMore" yaml:"zeroOrMore"`
}

func (n Node) String() string {
	switch n.Kind {
	case NodeAny:
		return n.Any.String()
	case NodeCharClass:
		return n.CharClass.String()
	case NodeConcat:
		return n.Concat.String()
	case NodeLiteral:
		return n.Literal.String()
	case NodeOneOrMore:
		return n.OneOrMore.String()
	case NodeOptional:
		return n.Optional.String()
	case NodeOr:
		return n.Or.String()
	case NodeRepetition:
		return n.Repetition.String()
	case NodeZeroOrMore:
		return n.ZeroOrMore.String()
	default:
		return "unknown"
	}
}

func (n *Node) IsSingleNode() bool {
	switch n.Kind {
	case NodeAny:
		return n.Any.IsSingleNode()
	case NodeCharClass:
		return n.CharClass.IsSingleNode()
	case NodeConcat:
		return n.Concat.IsSingleNode()
	case NodeLiteral:
		return n.Literal.IsSingleNode()
	case NodeOneOrMore:
		return n.OneOrMore.IsSingleNode()
	case NodeOptional:
		return n.Optional.IsSingleNode()
	case NodeOr:
		return n.Or.IsSingleNode()
	case NodeRepetition:
		return n.Repetition.IsSingleNode()
	case NodeZeroOrMore:
		return n.ZeroOrMore.IsSingleNode()
	default:
		return false
	}
}

func (n *Node) Validate() error {
	switch n.Kind {
	case NodeAny:
		return n.Any.Validate()
	case NodeCharClass:
		return n.CharClass.Validate()
	case NodeConcat:
		return n.Concat.Validate()
	case NodeLiteral:
		return n.Literal.Validate()
	case NodeOneOrMore:
		return n.OneOrMore.Validate()
	case NodeOptional:
		return n.Optional.Validate()
	case NodeOr:
		return n.Or.Validate()
	case NodeRepetition:
		return n.Repetition.Validate()
	case NodeZeroOrMore:
		return n.ZeroOrMore.Validate()
	default:
		return errors.New("unknown node kind")
	}
}

func (n Node) MarshalJSON() ([]byte, error) {

}

func (n *Node) UnmarshalJSON(data []byte) error {

}

func (n Node) MarshalJYAML() ([]byte, error) {

}

func (n *Node) UnmarshalYAML(data []byte) error {

}
