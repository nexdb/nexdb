package cache

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nexdb/nexdb/pkg/document"
)

// Operator is the type of operator to use in a condition.
type Operator string

const (
	Equals   Operator = "equals"
	Contains Operator = "contains"
)

// Condition is a condition to use in a query.
type Condition struct {
	Field    string   `json:"field"`
	Operator Operator `json:"operator"`
	Value    string   `json:"value"`
}

type Query struct {
	And []Element `json:"and,omitempty"`
	Or  []Element `json:"or,omitempty"`
}

type Element struct {
	Condition *Condition
	Query     *Query
}

// UnmarshalJSON overrides the default UnmarshalJSON for Element
func (e *Element) UnmarshalJSON(data []byte) error {
	// Try to unmarshal to Condition
	var cond Condition
	if err := json.Unmarshal(data, &cond); err == nil {
		e.Condition = &cond
		return nil
	}

	var query Query
	if err := json.Unmarshal(data, &query); err == nil {
		e.Query = &query
		return nil
	}

	return fmt.Errorf("could not unmarshal Element")
}

// Filter filters documents in the cache.
func (c *Cache) Filter(collection string, query Query) []*document.Document {
	c.RLock()
	defer c.RUnlock()
	results := []*document.Document{}
	for _, doc := range c.documents {
		if doc.Collection != collection {
			continue
		}

		if applyQuery(doc, query) {
			results = append(results, doc)
		}
	}
	return results
}

// getValueFromMap gets a value from a map using a dot-separated key.
func getValueFromMap(m map[string]interface{}, key string) interface{} {
	keys := strings.Split(key, ".")
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			return nil
		}
		if vm, ok := v.(map[string]interface{}); ok {
			m = vm
		} else {
			return v
		}
	}
	return nil
}

// applyCondition applies a condition to a single document.
func applyCondition(doc *document.Document, cond Condition) bool {
	switch cond.Operator {
	case Equals:
		return getValueFromMap(doc.Data, cond.Field) == cond.Value
	case Contains:
		str, ok := getValueFromMap(doc.Data, cond.Field).(string)
		return ok && strings.Contains(str, cond.Value)
	default:
		return false
	}
}

func applyElement(doc *document.Document, elem Element) bool {
	if elem.Condition != nil {
		return applyCondition(doc, *elem.Condition)
	}
	if elem.Query != nil {
		return applyQuery(doc, *elem.Query)
	}
	return false
}

func applyQuery(doc *document.Document, query Query) bool {
	if len(query.And) > 0 {
		for _, elem := range query.And {
			if !applyElement(doc, elem) {
				return false
			}
		}
	}

	if len(query.Or) > 0 {
		for _, elem := range query.Or {
			if applyElement(doc, elem) {
				return true
			}
		}
		// if there were OR conditions, we haven't found a match, so return false
		return false
	}
	// if there were no AND or OR conditions, return true
	return true
}
