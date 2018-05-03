package grules

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// OperatorAnd is what identifies the AND condition in a composite
	OperatorAnd = "and"
	// OperatorOr is what identifies the OR condition in a composite
	OperatorOr = "or"
)

// defaultComparators is a map of all the default comparators that
// a new engine should include
var defaultComparators = map[string]Comparator{
	"eq":        equal,
	"neq":       notEqual,
	"gt":        greaterThan,
	"gte":       greaterThanEqual,
	"lt":        lessThan,
	"lte":       lessThanEqual,
	"contains":  contains,
	"ncontains": notContains,
	"oneof":     oneOf,
}

// Rule is a our smallest unit of measure, each rule will be
// evaluated separately. The comparator is the logical operation to be
// performed, the path is the path into a map, delimited by '.', and
// the value is the value that we expect to match the value at the
// path
type Rule struct {
	Comparator string      `json:"comparator"`
	Path       string      `json:"path"`
	Value      interface{} `json:"value"`
}

// Composite is a group of rules that are joined by a logical operator
// AND or OR. If the operator is AND all of the rules must be true,
// if the operator is OR, one of the rules must be true.
type Composite struct {
	Operator   string      `json:"operator"`
	Rules      []Rule      `json:"rules"`
	Composites []Composite `json:"composites"`
}

// Engine is a group of composites. All of the composites must be
// true for the engine's evaluate function to return true.
type Engine struct {
	Composites  []Composite `json:"composites"`
	comparators map[string]Comparator
}

// NewEngine will create a new engine with the default comparators
func NewEngine() Engine {
	e := Engine{
		comparators: defaultComparators,
	}
	return e
}

// NewJSONEngine will create a new engine from it's JSON representation
func NewJSONEngine(raw json.RawMessage) (Engine, error) {
	var e Engine
	err := json.Unmarshal(raw, &e)
	if err != nil {
		return Engine{}, err
	}
	e.comparators = defaultComparators
	return e, nil
}

// AddComparator will add a new comparator that can be used in the
// engine's evaluation
func (e Engine) AddComparator(name string, c Comparator) Engine {
	e.comparators[name] = c
	return e
}

// Evaluate will ensure all of the composites in the engine are true
func (e Engine) Evaluate(props map[string]interface{}) bool {
	for _, c := range e.Composites {
		res := c.evaluate(props, e.comparators)
		if res == false {
			return false
		}
	}
	return true
}

// Stringify will generate a human readable rule set
func (e Engine) Stringify() string {
	parts := []string{}
	for _, c := range e.Composites {
		parts = append(parts, c.stringify(e.comparators))
	}

	return strings.Join(parts, " && ")
}

// Evaluate will ensure all either all of the rules are true, if given
// the AND operator, or that one of the rules is true if given the OR
// operator.
func (c Composite) evaluate(props map[string]interface{}, comps map[string]Comparator) bool {
	switch c.Operator {
	case OperatorAnd:
		for _, r := range c.Rules {
			res := r.evaluate(props, comps)
			if res == false {
				return false
			}
		}
		for _, cc := range c.Composites {
			res := cc.evaluate(props, comps)
			if res == false {
				return false
			}
		}
		return true
	case OperatorOr:
		for _, r := range c.Rules {
			res := r.evaluate(props, comps)
			if res == true {
				return true
			}
		}
		for _, cc := range c.Composites {
			res := cc.evaluate(props, comps)
			if res == true {
				return true
			}
		}
		return false
	}

	return false
}

// Stringify will generate a human readable rule set
func (c Composite) stringify(comps map[string]Comparator) string {
	s := "("
	parts := []string{}
	for _, r := range c.Rules {
		parts = append(parts, fmt.Sprintf("{%s %s %v}", r.Path, r.Comparator, r.Value))
	}
	for _, cc := range c.Composites {
		parts = append(parts, cc.stringify(comps))
	}
	s += strings.Join(parts, " "+c.Operator+" ")

	s += ")"
	return s
}

// Evaluate will return true if the rule is true, false otherwise
func (r Rule) evaluate(props map[string]interface{}, comps map[string]Comparator) bool {
	// Make sure we can get a value from the props
	val := pluck(props, r.Path)
	if val == nil {
		return false
	}

	comp, ok := comps[r.Comparator]
	if !ok {
		return false
	}

	return comp(val, r.Value)
}
