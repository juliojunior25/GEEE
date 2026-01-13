package jsontransformer

import (
	"testing"

	"github.com/yourusername/geee/pkg/types"
)

func BenchmarkPlugin_Execute_SingleMapping(b *testing.B) {
	plugin := New()

	ctx := &types.ExecutionContext{
		ExecutionID: "bench",
		State: map[string]interface{}{
			"input_field": "test_value",
		},
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"from": "input_field",
					"to":   "output_field",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := plugin.Execute(ctx)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}

func BenchmarkPlugin_Execute_MultipleMappings(b *testing.B) {
	plugin := New()

	ctx := &types.ExecutionContext{
		ExecutionID: "bench",
		State: map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
			"field3": "value3",
			"field4": "value4",
			"field5": "value5",
		},
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{"from": "field1", "to": "out1"},
				map[string]interface{}{"from": "field2", "to": "out2"},
				map[string]interface{}{"from": "field3", "to": "out3"},
				map[string]interface{}{"from": "field4", "to": "out4"},
				map[string]interface{}{"from": "field5", "to": "out5"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := plugin.Execute(ctx)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}

func BenchmarkPlugin_Execute_LargeState(b *testing.B) {
	plugin := New()

	// Create large state with 100 fields
	state := make(map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		state[string(rune('a'+i%26))+string(rune('0'+i/26))] = "value"
	}

	ctx := &types.ExecutionContext{
		ExecutionID: "bench",
		State:       state,
		Config: map[string]interface{}{
			"mappings": []interface{}{
				map[string]interface{}{
					"from": "a0",
					"to":   "output",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := plugin.Execute(ctx)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}
