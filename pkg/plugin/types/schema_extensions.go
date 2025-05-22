package types

import (
	"encoding/json"
	"fmt"
)

// XSourceField represents a field with an x-source extension.
type XSourceField struct {
	Name     string
	XSource  string
	Title    string
	Type     string
	Required bool
}

// ExtractXSourceFields parses a JSON schema and returns all fields with x-source extensions.
func ExtractXSourceFields(schemaData []byte) ([]XSourceField, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil, nil
	}
	// Get required fields
	requiredFields := map[string]bool{}
	if reqList, ok := schema["required"].([]interface{}); ok {
		for _, r := range reqList {
			if s, ok := r.(string); ok {
				requiredFields[s] = true
			}
		}
	}
	var result []XSourceField
	for name, prop := range props {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			continue
		}
		xSource, ok := propMap["x-source"].(string)
		if !ok {
			continue
		}
		title, _ := propMap["title"].(string)
		typ, _ := propMap["type"].(string)
		result = append(result, XSourceField{
			Name:     name,
			XSource:  xSource,
			Title:    title,
			Type:     typ,
			Required: requiredFields[name],
		})
	}
	return result, nil
}

// ValidateXSourceValue checks if a value is valid for a given x-source field.
func ValidateXSourceValue(field XSourceField, value string, fetcher func(xSource string) []string) bool {
	options := fetcher(field.XSource)
	for _, opt := range options {
		if opt == value {
			return true
		}
	}
	return false
}
