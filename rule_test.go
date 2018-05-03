package grules

import (
	"testing"
)

func TestRuleEvaluate(t *testing.T) {
	comparators := map[string]Comparator{
		"eq": equal,
	}
	props := map[string]interface{}{
		"first_name": "Trevor",
	}
	t.Run("basic rule", func(t *testing.T) {
		r := Rule{
			Comparator: "eq",
			Path:       "first_name",
			Value:      "Trevor",
		}
		res := r.evaluate(props, comparators)
		if res != true {
			t.Fatal("expected rule to be true")
		}
	})

	t.Run("unknown path", func(t *testing.T) {
		r := Rule{
			Comparator: "eq",
			Path:       "email",
			Value:      "Trevor",
		}
		res := r.evaluate(props, comparators)
		if res != false {
			t.Fatal("expected rule to be false")
		}
	})

	t.Run("non comparable types", func(t *testing.T) {
		r := Rule{
			Comparator: "eq",
			Path:       "name",
			Value:      func() {},
		}
		res := r.evaluate(props, comparators)
		if res != false {
			t.Fatal("expected rule to be false")
		}
	})

	t.Run("unknown comparator", func(t *testing.T) {
		r := Rule{
			Comparator: "unknown",
			Path:       "name",
			Value:      "Trevor",
		}
		res := r.evaluate(props, comparators)
		if res != false {
			t.Fatal("expected rule to be false")
		}
	})
}

func TestCompositeEvaluate(t *testing.T) {
	comparators := map[string]Comparator{
		"eq": equal,
		"gt": greaterThan,
		"lt": lessThan,
	}
	props := map[string]interface{}{
		"name": "Trevor",
		"age":  float64(23),
	}

	t.Run("and", func(t *testing.T) {
		c := Composite{
			Operator: OperatorAnd,
			Rules: []Rule{
				Rule{
					Comparator: "eq",
					Path:       "name",
					Value:      "Trevor",
				},
				Rule{
					Comparator: "eq",
					Path:       "age",
					Value:      float64(23),
				},
			},
		}
		res := c.evaluate(props, comparators)
		if res != true {
			t.Fatal("expected composite to be true")
		}
	})

	t.Run("or", func(t *testing.T) {
		c := Composite{
			Operator: OperatorOr,
			Rules: []Rule{
				Rule{
					Comparator: "eq",
					Path:       "name",
					Value:      "John",
				},
				Rule{
					Comparator: "eq",
					Path:       "age",
					Value:      float64(23),
				},
			},
		}
		res := c.evaluate(props, comparators)
		if res != true {
			t.Fatal("expected composite to be true")
		}
	})

	t.Run("nested and - or", func(t *testing.T) {
		c := Composite{
			Operator: OperatorAnd,
			Rules: []Rule{
				Rule{
					Comparator: "eq",
					Path:       "name",
					Value:      "Trevor",
				},
			},
			Composites: []Composite{
				Composite{
					Operator: OperatorOr,
					Rules: []Rule{
						Rule{
							Comparator: "gt",
							Path:       "age",
							Value:      float64(20),
						},
						Rule{
							Comparator: "lt",
							Path:       "age",
							Value:      float64(20),
						},
					},
				},
			},
		}
		res := c.evaluate(props, comparators)
		if res != true {
			t.Fatal("expected composite to be true")
		}
	})

	t.Run("nested or - and", func(t *testing.T) {
		c := Composite{
			Operator: OperatorOr,
			Rules: []Rule{
				Rule{
					Comparator: "eq",
					Path:       "name",
					Value:      "John",
				},
			},
			Composites: []Composite{
				Composite{
					Operator: OperatorAnd,
					Rules: []Rule{
						Rule{
							Comparator: "gt",
							Path:       "age",
							Value:      float64(20),
						},
						Rule{
							Comparator: "lt",
							Path:       "age",
							Value:      float64(30),
						},
					},
				},
			},
		}
		res := c.evaluate(props, comparators)
		if res != true {
			t.Fatal("expected composite to be true")
		}
	})

	t.Run("unknown operator", func(t *testing.T) {
		c := Composite{
			Operator: "unknown",
			Rules: []Rule{
				Rule{
					Comparator: "eq",
					Path:       "name",
					Value:      "John",
				},
				Rule{
					Comparator: "eq",
					Path:       "age",
					Value:      float64(23),
				},
			},
		}
		res := c.evaluate(props, comparators)
		if res != false {
			t.Fatal("expected composite to be true")
		}
	})
}

func TestAddComparator(t *testing.T) {
	comp := func(a, b interface{}) bool {
		return false
	}
	e := NewEngine()
	e = e.AddComparator("always-false", comp)
	if e.comparators["always-false"] == nil {
		t.Fatal("expected comparator to be added under key always-false")
	}

	e.Composites = []Composite{
		Composite{
			Operator: OperatorAnd,
			Rules: []Rule{
				Rule{
					Comparator: "always-false",
					Path:       "user.name",
					Value:      "Trevor",
				},
			},
		},
	}

	props := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Trevor",
		},
	}

	res := e.Evaluate(props)
	if res != false {
		t.Fatal("expected engine to be false")
	}
}

func TestNewJSONEngine(t *testing.T) {
	j := []byte(`{"composites":[{"operator":"and","rules":[{"comparator":"eq","path":"first_name","value":"Trevor"}]}]}`)
	e, err := NewJSONEngine(j)
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Composites) != 1 {
		t.Fatal("expected 1 composite")
	}
	if len(e.Composites[0].Rules) != 1 {
		t.Fatal("expected 1 rule in first composite")
	}
}

func TestEngineEvaluate(t *testing.T) {
	t.Run("no composites", func(t *testing.T) {
		props := map[string]interface{}{
			"user": map[string]interface{}{
				"email": "test@test.com",
				"name":  "Trevor",
				"id":    float64(1234),
			},
		}
		e := NewEngine()
		res := e.Evaluate(props)
		if res != true {
			t.Fatal("expected engine to pass")
		}
		humanReadable := e.Stringify()
		t.Log(humanReadable)
	})

	t.Run("1 composite, 1 rule", func(t *testing.T) {
		props := map[string]interface{}{
			"address": map[string]interface{}{
				"bedroom": map[string]interface{}{
					"furniture": []interface{}{
						"bed",
						"tv",
						"dresser",
					},
				},
			},
		}
		e := NewEngine()
		e.Composites = []Composite{
			Composite{
				Operator: OperatorAnd,
				Rules: []Rule{
					Rule{
						Comparator: "contains",
						Path:       "address.bedroom.furniture",
						Value:      "tv",
					},
				},
			},
		}
		res := e.Evaluate(props)
		if res != true {
			t.Fatal("expected engine to pass")
		}

		expectedStr := "({address.bedroom.furniture contains tv})"
		actualStr := e.Stringify()
		if expectedStr != actualStr {
			t.Fatalf("expected %s but got %s", expectedStr, actualStr)
		}
	})

	t.Run("2 composites, 1 rule", func(t *testing.T) {
		props := map[string]interface{}{
			"user": map[string]interface{}{
				"email": "test@test.com",
				"name":  "Trevor",
				"id":    float64(1234),
			},
		}
		e := NewEngine()
		e.Composites = []Composite{
			Composite{
				Operator: OperatorAnd,
				Rules: []Rule{
					Rule{
						Comparator: "eq",
						Path:       "user.name",
						Value:      "Trevor",
					},
					Rule{
						Comparator: "eq",
						Path:       "user.id",
						Value:      float64(1234),
					},
				},
			},
			Composite{
				Operator: OperatorOr,
				Rules: []Rule{
					Rule{
						Comparator: "eq",
						Path:       "user.name",
						Value:      "Trevor",
					},
					Rule{
						Comparator: "eq",
						Path:       "user.id",
						Value:      float64(7),
					},
				},
			},
		}
		res := e.Evaluate(props)
		if res != true {
			t.Fatal("expected engine to pass")
		}

		expectedStr := "({user.name eq Trevor} and {user.id eq 1234}) && ({user.name eq Trevor} or {user.id eq 7})"
		actualStr := e.Stringify()
		if expectedStr != actualStr {
			t.Fatalf("expected %s but got %s", expectedStr, actualStr)
		}
	})
}
