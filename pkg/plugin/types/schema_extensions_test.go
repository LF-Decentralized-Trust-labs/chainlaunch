package types

import (
	"testing"
)

func TestExtractXSourceFields(t *testing.T) {
	schema := []byte(`{
    "type": "object",
    "properties": {
      "KEY_ID": {
        "type": "string",
        "title": "Private Key",
        "x-source": "keyStore"
      },
      "FABRIC_ORG": {
        "type": "string",
        "title": "Fabric Org",
        "x-source": "fabricOrgs"
      },
      "PLAIN": {
        "type": "string",
        "title": "Plain"
      }
    },
    "required": ["KEY_ID"]
  }`)
	fields, err := ExtractXSourceFields(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 x-source fields, got %d", len(fields))
	}
	if fields[0].Name != "KEY_ID" || fields[0].XSource != "keyStore" || !fields[0].Required {
		t.Errorf("unexpected field: %+v", fields[0])
	}
	if fields[1].Name != "FABRIC_ORG" || fields[1].XSource != "fabricOrgs" || fields[1].Required {
		t.Errorf("unexpected field: %+v", fields[1])
	}
}

func TestValidateXSourceValue(t *testing.T) {
	field := XSourceField{Name: "KEY_ID", XSource: "keyStore"}
	fetcher := func(xSource string) []string {
		if xSource == "keyStore" {
			return []string{"key1", "key2"}
		}
		return nil
	}
	if !ValidateXSourceValue(field, "key1", fetcher) {
		t.Error("expected key1 to be valid")
	}
	if ValidateXSourceValue(field, "key3", fetcher) {
		t.Error("expected key3 to be invalid")
	}
}
