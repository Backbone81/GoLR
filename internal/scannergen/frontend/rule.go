package frontend

// Rule describes a single token with name and regular expression.
type Rule struct {
	Name  string `json:"name"           yaml:"name"`
	Regex Node   `json:"regex"          yaml:"regex"`
	Skip  bool   `json:"skip,omitempty" yaml:"skip,omitempty"`
}
