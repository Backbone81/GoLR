package frontend

import (
	"encoding/json"
	"errors"

	"github.com/goccy/go-yaml"
)

// Kind is a type describing the kind of regular expression node.
type Kind int

const (
	KindAny Kind = iota
	KindCharClass
	KindConcat
	KindLiteral
	KindOneOrMore
	KindOptional
	KindOr
	KindRepetition
	KindZeroOrMore
)

// Node is a single node of a regular expression.
type Node struct {
	Kind Kind `json:"kind" yaml:"kind"`

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
	case KindAny:
		return n.Any.String()
	case KindCharClass:
		return n.CharClass.String()
	case KindConcat:
		return n.Concat.String()
	case KindLiteral:
		return n.Literal.String()
	case KindOneOrMore:
		return n.OneOrMore.String()
	case KindOptional:
		return n.Optional.String()
	case KindOr:
		return n.Or.String()
	case KindRepetition:
		return n.Repetition.String()
	case KindZeroOrMore:
		return n.ZeroOrMore.String()
	default:
		return "unknown"
	}
}

func (n *Node) IsSingleNode() bool {
	switch n.Kind {
	case KindAny:
		return n.Any.IsSingleNode()
	case KindCharClass:
		return n.CharClass.IsSingleNode()
	case KindConcat:
		return n.Concat.IsSingleNode()
	case KindLiteral:
		return n.Literal.IsSingleNode()
	case KindOneOrMore:
		return n.OneOrMore.IsSingleNode()
	case KindOptional:
		return n.Optional.IsSingleNode()
	case KindOr:
		return n.Or.IsSingleNode()
	case KindRepetition:
		return n.Repetition.IsSingleNode()
	case KindZeroOrMore:
		return n.ZeroOrMore.IsSingleNode()
	default:
		return false
	}
}

func (n *Node) Validate() error {
	switch n.Kind {
	case KindAny:
		return n.Any.Validate()
	case KindCharClass:
		return n.CharClass.Validate()
	case KindConcat:
		return n.Concat.Validate()
	case KindLiteral:
		return n.Literal.Validate()
	case KindOneOrMore:
		return n.OneOrMore.Validate()
	case KindOptional:
		return n.Optional.Validate()
	case KindOr:
		return n.Or.Validate()
	case KindRepetition:
		return n.Repetition.Validate()
	case KindZeroOrMore:
		return n.ZeroOrMore.Validate()
	default:
		return errors.New("unknown node kind")
	}
}

// marshalNode is a helper struct for marshaling to JSON or YAML. We don't want to marshal the structs for all kinds.
// We only want to marshal that struct for the correct kind. Therefore, we need pointers with omitempty in this helper
// struct while keeping the nested values in the primary struct.
type marshalNode struct {
	Kind Kind `json:"kind" yaml:"kind"`

	CharClass  *CharClass  `json:"charClass,omitempty" yaml:"charClass,omitempty"`
	Concat     *Concat     `json:"concat,omitempty" yaml:"concat,omitempty"`
	Literal    *Literal    `json:"literal,omitempty" yaml:"literal,omitempty"`
	OneOrMore  *OneOrMore  `json:"oneOrMore,omitempty" yaml:"oneOrMore,omitempty"`
	Optional   *Optional   `json:"optional,omitempty" yaml:"optional,omitempty"`
	Or         *Or         `json:"or,omitempty" yaml:"or,omitempty"`
	Repetition *Repetition `json:"repetition,omitempty" yaml:"repetition,omitempty"`
	ZeroOrMore *ZeroOrMore `json:"zeroOrMore,omitempty" yaml:"zeroOrMore,omitempty"`
}

func (n Node) MarshalJSON() ([]byte, error) {
	node := marshalNode{
		Kind: n.Kind,
	}

	switch n.Kind {
	case KindAny:
		// nothing to do
	case KindCharClass:
		node.CharClass = &n.CharClass
	case KindConcat:
		node.Concat = &n.Concat
	case KindLiteral:
		node.Literal = &n.Literal
	case KindOneOrMore:
		node.OneOrMore = &n.OneOrMore
	case KindOptional:
		node.Optional = &n.Optional
	case KindOr:
		node.Or = &n.Or
	case KindRepetition:
		node.Repetition = &n.Repetition
	case KindZeroOrMore:
		node.ZeroOrMore = &n.ZeroOrMore
	default:
		return nil, errors.New("unknown node kind")
	}
	return json.Marshal(node)
}

func (n *Node) UnmarshalJSON(data []byte) error {
	var node marshalNode
	if err := json.Unmarshal(data, &node); err != nil {
		return err
	}

	n.Kind = node.Kind
	switch n.Kind {
	case KindAny:
		// nothing to do
	case KindCharClass:
		n.CharClass = *node.CharClass
	case KindConcat:
		n.Concat = *node.Concat
	case KindLiteral:
		n.Literal = *node.Literal
	case KindOneOrMore:
		n.OneOrMore = *node.OneOrMore
	case KindOptional:
		n.Optional = *node.Optional
	case KindOr:
		n.Or = *node.Or
	case KindRepetition:
		n.Repetition = *node.Repetition
	case KindZeroOrMore:
		n.ZeroOrMore = *node.ZeroOrMore
	default:
		return errors.New("unknown node kind")
	}
	return nil
}

func (n Node) MarshalYAML() ([]byte, error) {
	node := marshalNode{
		Kind: n.Kind,
	}

	switch n.Kind {
	case KindAny:
		// nothing to do
	case KindCharClass:
		node.CharClass = &n.CharClass
	case KindConcat:
		node.Concat = &n.Concat
	case KindLiteral:
		node.Literal = &n.Literal
	case KindOneOrMore:
		node.OneOrMore = &n.OneOrMore
	case KindOptional:
		node.Optional = &n.Optional
	case KindOr:
		node.Or = &n.Or
	case KindRepetition:
		node.Repetition = &n.Repetition
	case KindZeroOrMore:
		node.ZeroOrMore = &n.ZeroOrMore
	default:
		return nil, errors.New("unknown node kind")
	}
	return yaml.Marshal(node)
}

func (n *Node) UnmarshalYAML(data []byte) error {
	var node marshalNode
	if err := yaml.Unmarshal(data, &node); err != nil {
		return err
	}

	n.Kind = node.Kind
	switch n.Kind {
	case KindAny:
		// nothing to do
	case KindCharClass:
		n.CharClass = *node.CharClass
	case KindConcat:
		n.Concat = *node.Concat
	case KindLiteral:
		n.Literal = *node.Literal
	case KindOneOrMore:
		n.OneOrMore = *node.OneOrMore
	case KindOptional:
		n.Optional = *node.Optional
	case KindOr:
		n.Or = *node.Or
	case KindRepetition:
		n.Repetition = *node.Repetition
	case KindZeroOrMore:
		n.ZeroOrMore = *node.ZeroOrMore
	default:
		return errors.New("unknown node kind")
	}
	return nil
}
