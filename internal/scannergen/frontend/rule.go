package frontend

import "fmt"

// Rule describes a single token with name and regular expression.
type Rule struct {
	Name  string `json:"name"           yaml:"name"`
	Regex Node   `json:"regex"          yaml:"regex"`
	Skip  bool   `json:"skip,omitempty" yaml:"skip,omitempty"`
}

// ValidateRules returns an error for the first rule with an invalid regular expression. Frontends which decode rules
// as they are call it, because the core expects valid regular expressions.
func ValidateRules(rules []Rule) error {
	for _, rule := range rules {
		if err := rule.Regex.Validate(); err != nil {
			return fmt.Errorf("rule %q: %w", rule.Name, err)
		}
	}
	return nil
}
