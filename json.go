package aslite

import (
	"encoding/json"
	"reflect"
)

func UnmarshalJSONWithExtra(data []byte, v interface{}) (map[string]interface{}, error) {
	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	t := rv.Type()
	fields := make(map[string]struct{})
	extractFields(t, fields, "", "json")

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	extra := make(map[string]interface{})
	for k, v := range raw {
		if _, ok := fields[k]; !ok {
			extra[k] = v
		}
	}

	return extra, nil
}

func extractFields(t reflect.Type, fields map[string]struct{}, prefix string, tagName string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			extractFields(field.Type, fields, prefix, tagName)
		} else {
			tag := field.Tag.Get(tagName)
			if tag == "" {
				tag = field.Name
			}
			fields[prefix+tag] = struct{}{}
		}
	}
}

// MarshalJSONWithExtra returns JSON bytes with extra fields
func MarshalJSONWithExtra(v interface{}, extra map[string]interface{}) ([]byte, error) {
	valueJSON, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var value map[string]interface{}
	if err := json.Unmarshal(valueJSON, &value); err != nil {
		return nil, err
	}
	for k, v := range extra {
		value[k] = v
	}

	return json.Marshal(value)
}
