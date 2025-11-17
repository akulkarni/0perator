package operator

import (
	"context"
	"testing"
	"time"

	"github.com/akulkarni/0perator/internal/actions"
)

func TestOperatorDiscovery(t *testing.T) {
	op := New()

	// Test discovering all actions
	allActions := op.GetAvailableActions("")
	if len(allActions) < 2 {
		t.Errorf("Expected at least 2 actions, got %d", len(allActions))
	}

	// Test discovering by query
	webActions := op.DiscoverActions("web")
	if len(webActions) == 0 {
		t.Error("Expected to find web-related actions")
	}

	// Test discovering by category
	createActions := op.GetAvailableActions("create")
	if len(createActions) == 0 {
		t.Error("Expected to find create category actions")
	}
}

func TestOperatorExecuteSingle(t *testing.T) {
	op := New()
	ctx := context.Background()

	// Create a test action
	testAction := &actions.Action{
		Name:         "test_action",
		Description:  "Test action for unit tests",
		Category:     actions.CategoryUtil,
		Tags:         []string{"test"},
		Tier:         actions.TierFast,
		EstimatedTime: 1 * time.Second,
		Inputs: []actions.Input{
			{
				Name:     "input1",
				Type:     actions.InputTypeString,
				Required: true,
			},
		},
		Outputs: []actions.Output{
			{
				Name: "output1",
				Type: actions.InputTypeString,
			},
		},
		Dependencies: []string{},
		Conflicts:    []string{},
		Implementation: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"output1": "result_" + inputs["input1"].(string),
			}, nil
		},
	}

	// Register test action
	op.registry.Register(testAction)

	// Execute the action
	result, err := op.ExecuteAction(ctx, "test_action", map[string]interface{}{
		"input1": "test",
	})

	if err != nil {
		t.Fatalf("Failed to execute action: %v", err)
	}

	if !result.Success {
		t.Errorf("Action execution failed: %s", result.Error)
	}

	if result.Outputs["output1"] != "result_test" {
		t.Errorf("Unexpected output: %v", result.Outputs["output1"])
	}
}

func TestOperatorExecuteSequence(t *testing.T) {
	op := New()
	ctx := context.Background()

	// Create test actions with dependencies
	action1 := &actions.Action{
		Name:         "step1",
		Description:  "First step",
		Category:     actions.CategoryUtil,
		Tier:         actions.TierFast,
		EstimatedTime: 1 * time.Second,
		Inputs:       []actions.Input{},
		Outputs: []actions.Output{
			{Name: "data1", Type: actions.InputTypeString},
		},
		Dependencies: []string{},
		Implementation: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"data1": "value1"}, nil
		},
	}

	action2 := &actions.Action{
		Name:         "step2",
		Description:  "Second step",
		Category:     actions.CategoryUtil,
		Tier:         actions.TierFast,
		EstimatedTime: 1 * time.Second,
		Inputs:       []actions.Input{},
		Outputs: []actions.Output{
			{Name: "data2", Type: actions.InputTypeString},
		},
		Dependencies: []string{"step1"},
		Implementation: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			// Can access outputs from previous actions through inputs
			return map[string]interface{}{"data2": "value2"}, nil
		},
	}

	// Register actions
	op.registry.Register(action1)
	op.registry.Register(action2)

	// Execute sequence (intentionally out of order to test dependency resolution)
	calls := []actions.ActionCall{
		{Action: "step2", Inputs: map[string]interface{}{}},
		{Action: "step1", Inputs: map[string]interface{}{}},
	}

	result, err := op.ExecuteSequence(ctx, calls)
	if err != nil {
		t.Fatalf("Failed to execute sequence: %v", err)
	}

	if !result.Success {
		t.Error("Sequence execution failed")
	}

	// Check that actions executed in correct order
	if len(result.Actions) != 2 {
		t.Errorf("Expected 2 action results, got %d", len(result.Actions))
	}

	// step1 should have executed before step2 due to dependency
	if result.Actions[0].Action != "step1" {
		t.Errorf("Expected step1 to execute first, got %s", result.Actions[0].Action)
	}

	if result.Actions[1].Action != "step2" {
		t.Errorf("Expected step2 to execute second, got %s", result.Actions[1].Action)
	}

	// Check outputs are combined
	if result.Outputs["data1"] != "value1" {
		t.Errorf("Missing or incorrect data1 output")
	}

	if result.Outputs["data2"] != "value2" {
		t.Errorf("Missing or incorrect data2 output")
	}
}

func TestOperatorValidation(t *testing.T) {
	op := New()

	// Test with non-existent action
	err := op.ValidateSequence([]actions.ActionCall{
		{Action: "non_existent", Inputs: map[string]interface{}{}},
	})

	if err == nil {
		t.Error("Expected error for non-existent action")
	}

	// Test with conflicting actions
	conflict1 := &actions.Action{
		Name:         "db1",
		Description:  "Database 1",
		Category:     actions.CategorySetup,
		Tier:         actions.TierFast,
		EstimatedTime: 1 * time.Second,
		Conflicts:    []string{"db2"},
		Implementation: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
	}

	conflict2 := &actions.Action{
		Name:         "db2",
		Description:  "Database 2",
		Category:     actions.CategorySetup,
		Tier:         actions.TierFast,
		EstimatedTime: 1 * time.Second,
		Conflicts:    []string{"db1"},
		Implementation: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
	}

	op.registry.Register(conflict1)
	op.registry.Register(conflict2)

	err = op.ValidateSequence([]actions.ActionCall{
		{Action: "db1", Inputs: map[string]interface{}{}},
		{Action: "db2", Inputs: map[string]interface{}{}},
	})

	if err == nil {
		t.Error("Expected error for conflicting actions")
	}
}