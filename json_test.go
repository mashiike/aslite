package aslite_test

import (
	"testing"

	"github.com/mashiike/aslite"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"Field2"`
}

func TestUnmarshalJSONWithExtra(t *testing.T) {
	jsonData := `{"field1": "value1", "Field2": 2, "extraField": "extraValue"}`

	var ts TestStruct
	extra, err := aslite.UnmarshalJSONWithExtra([]byte(jsonData), &ts)
	if err != nil {
		t.Fatalf("UnmarshalJSONWithExtra returned an error: %v", err)
	}
	require.EqualValues(t, TestStruct{Field1: "value1", Field2: 2}, ts, "value of ts")
	require.EqualValues(t, map[string]interface{}{"extraField": "extraValue"}, extra, "value of extra")
}

func TestUnmarshalJSONWithExtra__Alias(t *testing.T) {
	jsonData := `{"field1": "value1", "Field2": 2, "extraField": "extraValue"}`

	var ts TestStruct
	type TestStructAlias TestStruct
	aux := &struct {
		*TestStructAlias
	}{
		TestStructAlias: (*TestStructAlias)(&ts),
	}
	extra, err := aslite.UnmarshalJSONWithExtra([]byte(jsonData), aux)
	if err != nil {
		t.Fatalf("UnmarshalJSONWithExtra returned an error: %v", err)
	}
	require.EqualValues(t, TestStructAlias{Field1: "value1", Field2: 2}, ts, "value of ts")
	require.EqualValues(t, map[string]interface{}{"extraField": "extraValue"}, extra, "value of extra")
}

func TestMarshalJSONWithExtra(t *testing.T) {
	ts := TestStruct{Field1: "value1", Field2: 2}
	extra := map[string]interface{}{"extraField": "extraValue"}

	bs, err := aslite.MarshalJSONWithExtra(ts, extra)
	if err != nil {
		t.Fatalf("MarshalJSONWithExtra returned an error: %v", err)
	}
	require.JSONEq(t, `{"Field2":2,"extraField":"extraValue","field1":"value1"}`, string(bs), "value of bs")
}

func TestMarshalJSONWithExtra__Alias(t *testing.T) {
	ts := TestStruct{Field1: "value1", Field2: 2}
	extra := map[string]interface{}{"extraField": "extraValue"}

	type TestStructAlias TestStruct
	aux := &struct {
		*TestStructAlias
	}{
		TestStructAlias: (*TestStructAlias)(&ts),
	}

	bs, err := aslite.MarshalJSONWithExtra(aux, extra)
	if err != nil {
		t.Fatalf("MarshalJSONWithExtra returned an error: %v", err)
	}
	require.JSONEq(t, `{"Field2":2,"extraField":"extraValue","field1":"value1"}`, string(bs), "value of bs")
}
