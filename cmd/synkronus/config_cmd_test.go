package main

import (
	"testing"
)

func TestFlattenConfigMap_EmptyMap(t *testing.T) {
	result := flattenConfigMap(map[string]interface{}{})
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestFlattenConfigMap_SingleLevel(t *testing.T) {
	input := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	result := flattenConfigMap(input)
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %v", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("expected key2=value2, got %v", result["key2"])
	}
}

func TestFlattenConfigMap_NestedTwoLevels(t *testing.T) {
	input := map[string]interface{}{
		"gcp": map[string]interface{}{
			"project": "my-project",
		},
	}
	result := flattenConfigMap(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d: %v", len(result), result)
	}
	if result["gcp.project"] != "my-project" {
		t.Errorf("expected gcp.project=my-project, got %v", result["gcp.project"])
	}
}

func TestFlattenConfigMap_DeeplyNested(t *testing.T) {
	input := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "deep-value",
			},
		},
	}
	result := flattenConfigMap(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d: %v", len(result), result)
	}
	if result["a.b.c"] != "deep-value" {
		t.Errorf("expected a.b.c=deep-value, got %v", result["a.b.c"])
	}
}

func TestFlattenConfigMap_MixedValueTypes(t *testing.T) {
	input := map[string]interface{}{
		"str":  "hello",
		"num":  42,
		"flag": true,
	}
	result := flattenConfigMap(input)
	if len(result) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result))
	}
	if result["str"] != "hello" {
		t.Errorf("expected str=hello, got %v", result["str"])
	}
	if result["num"] != 42 {
		t.Errorf("expected num=42, got %v", result["num"])
	}
	if result["flag"] != true {
		t.Errorf("expected flag=true, got %v", result["flag"])
	}
}

func TestFlattenConfigMap_MultipleNestedBranches(t *testing.T) {
	input := map[string]interface{}{
		"gcp": map[string]interface{}{
			"project": "p1",
		},
		"aws": map[string]interface{}{
			"region": "us-east-1",
		},
	}
	result := flattenConfigMap(input)
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(result), result)
	}
	if result["gcp.project"] != "p1" {
		t.Errorf("expected gcp.project=p1, got %v", result["gcp.project"])
	}
	if result["aws.region"] != "us-east-1" {
		t.Errorf("expected aws.region=us-east-1, got %v", result["aws.region"])
	}
}

func TestFlattenConfigMap_NilValue(t *testing.T) {
	input := map[string]interface{}{
		"key": nil,
	}
	result := flattenConfigMap(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result["key"] != nil {
		t.Errorf("expected key=nil, got %v", result["key"])
	}
}
