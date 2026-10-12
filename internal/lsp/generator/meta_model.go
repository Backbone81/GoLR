package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MetaModel is the meta model of the Language Server Protocol as defined by metaModel.ts of the specification.
type MetaModel struct {
	MetaData      MetaData       `json:"metaData"`
	Requests      []Request      `json:"requests"`
	Notifications []Notification `json:"notifications"`
	Structures    []Structure    `json:"structures"`
	Enumerations  []Enumeration  `json:"enumerations"`
	TypeAliases   []TypeAlias    `json:"typeAliases"`
}

// MetaData describes the meta model itself.
type MetaData struct {
	// Version is the version of the specification, e.g. 3.18.0.
	Version string `json:"version"`
}

// Annotations are the properties every element of the meta model can carry.
type Annotations struct {
	Documentation string   `json:"documentation"`
	Since         string   `json:"since"`
	SinceTags     []string `json:"sinceTags"`
	Proposed      bool     `json:"proposed"`
	Deprecated    string   `json:"deprecated"`
}

// MessageDirection is the side which sends a message.
type MessageDirection string

const (
	MessageDirectionClientToServer MessageDirection = "clientToServer"
	MessageDirectionServerToClient MessageDirection = "serverToClient"
	MessageDirectionBoth           MessageDirection = "both"
)

// Request is a request of the protocol.
type Request struct {
	Annotations

	Method              string           `json:"method"`
	TypeName            string           `json:"typeName"`
	Params              *Type            `json:"params"`
	Result              *Type            `json:"result"`
	PartialResult       *Type            `json:"partialResult"`
	ErrorData           *Type            `json:"errorData"`
	RegistrationMethod  string           `json:"registrationMethod"`
	RegistrationOptions *Type            `json:"registrationOptions"`
	MessageDirection    MessageDirection `json:"messageDirection"`
	ClientCapability    string           `json:"clientCapability"`
	ServerCapability    string           `json:"serverCapability"`
}

// Notification is a notification of the protocol.
type Notification struct {
	Annotations

	Method              string           `json:"method"`
	TypeName            string           `json:"typeName"`
	Params              *Type            `json:"params"`
	RegistrationMethod  string           `json:"registrationMethod"`
	RegistrationOptions *Type            `json:"registrationOptions"`
	MessageDirection    MessageDirection `json:"messageDirection"`
	ClientCapability    string           `json:"clientCapability"`
	ServerCapability    string           `json:"serverCapability"`
}

// Property is a property of a structure or a structure literal.
type Property struct {
	Annotations

	Name     string `json:"name"`
	Type     Type   `json:"type"`
	Optional bool   `json:"optional"`
}

// Structure is a named structure.
type Structure struct {
	Annotations

	Name       string     `json:"name"`
	Extends    []Type     `json:"extends"`
	Mixins     []Type     `json:"mixins"`
	Properties []Property `json:"properties"`
}

// StructureLiteral is a structure without a name, declared in place.
type StructureLiteral struct {
	Annotations

	Properties []Property `json:"properties"`
}

// TypeAlias is a name for a type.
type TypeAlias struct {
	Annotations

	Name string `json:"name"`
	Type Type   `json:"type"`
}

// Enumeration is a named set of values of a base type.
type Enumeration struct {
	Annotations

	Name                 string             `json:"name"`
	Type                 Type               `json:"type"`
	Values               []EnumerationEntry `json:"values"`
	SupportsCustomValues bool               `json:"supportsCustomValues"`
}

// EnumerationEntry is a value of an enumeration.
type EnumerationEntry struct {
	Annotations

	Name string `json:"name"`

	// Value is the JSON string or number of the entry.
	Value json.RawMessage `json:"value"`
}

// ParseMetaModel decodes the JSON of a meta model. Properties the meta model does not define are an error, so a newer
// meta model fails instead of losing information.
func ParseMetaModel(data []byte) (*MetaModel, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var model MetaModel
	if err := decoder.Decode(&model); err != nil {
		return nil, fmt.Errorf("decoding the meta model: %w", err)
	}
	return &model, nil
}
