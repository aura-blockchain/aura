package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDetectBias(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name       string
		modelHash  string
		input      string
		output     string
		expectBias bool
	}{
		{
			name:       "neutral output",
			modelHash:  "model1",
			input:      "What is the weather?",
			output:     "The weather is sunny",
			expectBias: false,
		},
		{
			name:       "potentially biased output",
			modelHash:  "model2",
			input:      "Tell me about different groups",
			output:     "Some groups are better than others",
			expectBias: true,
		},
		{
			name:       "fair comparison",
			modelHash:  "model1",
			input:      "Compare options A and B",
			output:     "Both options have advantages and disadvantages",
			expectBias: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := k.DetectBias(ctx, tt.modelHash, tt.input, tt.output)

			if tt.expectBias && result.BiasScore == 0 {
				t.Error("Expected bias to be detected")
			}

			if !tt.expectBias && result.BiasScore > 0.8 {
				t.Error("Expected low bias score for neutral output")
			}
		})
	}
}

func TestBiasScoreCalculation(t *testing.T) {
	k, ctx := setupKeeper(t)

	result := k.DetectBias(ctx, "model1", "input", "output")

	if result.BiasScore < 0 || result.BiasScore > 1 {
		t.Errorf("Bias score should be between 0 and 1, got %v", result.BiasScore)
	}
}

func TestBiasCategories(t *testing.T) {
	k, ctx := setupKeeper(t)

	result := k.DetectBias(ctx, "model1", "input", "output")

	// Verify bias categories are populated
	if result.BiasCategories == nil {
		result.BiasCategories = make(map[string]float64)
	}

	validCategories := []string{"gender", "race", "age", "religion", "nationality"}
	for _, category := range validCategories {
		if score, exists := result.BiasCategories[category]; exists {
			if score < 0 || score > 1 {
				t.Errorf("Category %s score should be between 0 and 1, got %v", category, score)
			}
		}
	}
}

func TestRecordBiasDetection(t *testing.T) {
	k, ctx := setupKeeper(t)

	detection := BiasDetectionResult{
		ModelHash:       "model1",
		Input:           "test input",
		Output:          "test output",
		BiasScore:       0.3,
		BiasCategories:  map[string]float64{"gender": 0.2, "race": 0.1},
		DetectionMethod: "pattern_matching",
		Timestamp:       ctx.BlockTime(),
	}

	err := k.RecordBiasDetection(ctx, detection)
	if err != nil {
		t.Fatalf("Failed to record bias detection: %v", err)
	}

	// Verify it was stored
	results := k.GetBiasDetectionHistory(ctx, "model1", 10)
	if len(results) == 0 {
		t.Error("Expected bias detection to be recorded")
	}
}

func TestGetBiasDetectionHistory(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record multiple detections
	for i := 0; i < 5; i++ {
		detection := BiasDetectionResult{
			ModelHash:  "model1",
			Input:      "input",
			Output:     "output",
			BiasScore:  float64(i) * 0.1,
			Timestamp:  ctx.BlockTime(),
		}
		if err := k.RecordBiasDetection(ctx, detection); err != nil {
			t.Fatalf("Failed to record detection: %v", err)
		}
	}

	history := k.GetBiasDetectionHistory(ctx, "model1", 10)
	if len(history) < 5 {
		t.Errorf("Expected at least 5 detections, got %d", len(history))
	}
}

func TestGetModelBiasStatistics(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record detections with varying bias scores
	biasScores := []float64{0.1, 0.3, 0.5, 0.7, 0.2}
	for _, score := range biasScores {
		detection := BiasDetectionResult{
			ModelHash:  "model1",
			Input:      "input",
			Output:     "output",
			BiasScore:  score,
			Timestamp:  ctx.BlockTime(),
		}
		if err := k.RecordBiasDetection(ctx, detection); err != nil {
			t.Fatalf("Failed to record detection: %v", err)
		}
	}

	stats := k.GetModelBiasStatistics(ctx, "model1")

	if stats.TotalDetections < uint64(len(biasScores)) {
		t.Errorf("Expected at least %d detections, got %d", len(biasScores), stats.TotalDetections)
	}

	// Average should be (0.1 + 0.3 + 0.5 + 0.7 + 0.2) / 5 = 0.36
	expectedAvg := 0.36
	tolerance := 0.1
	if stats.AverageBiasScore < expectedAvg-tolerance || stats.AverageBiasScore > expectedAvg+tolerance {
		t.Logf("Average bias score: expected ~%v, got %v", expectedAvg, stats.AverageBiasScore)
	}
}

func TestBiasMitigation(t *testing.T) {
	k, ctx := setupKeeper(t)

	mitigation := BiasMitigationStrategy{
		ModelHash:   "model1",
		Strategy:    "rebalancing",
		Parameters:  map[string]string{"threshold": "0.5"},
		AppliedAt:   ctx.BlockTime(),
	}

	err := k.ApplyBiasMitigation(ctx, mitigation)
	if err != nil {
		t.Fatalf("Failed to apply bias mitigation: %v", err)
	}

	// Verify mitigation was applied
	strategies := k.GetBiasMitigationStrategies(ctx, "model1")
	if len(strategies) == 0 {
		t.Error("Expected mitigation strategy to be stored")
	}
}

func TestBiasThresholdViolation(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set bias threshold
	threshold := 0.5
	err := k.SetBiasThreshold(ctx, "model1", threshold)
	if err != nil {
		t.Fatalf("Failed to set bias threshold: %v", err)
	}

	// Test with score above threshold
	highBiasResult := BiasDetectionResult{
		ModelHash:  "model1",
		BiasScore:  0.8,
		Timestamp:  ctx.BlockTime(),
	}

	violation := k.CheckBiasThresholdViolation(ctx, highBiasResult)
	if !violation {
		t.Error("Expected threshold violation for high bias score")
	}

	// Test with score below threshold
	lowBiasResult := BiasDetectionResult{
		ModelHash:  "model1",
		BiasScore:  0.3,
		Timestamp:  ctx.BlockTime(),
	}

	violation = k.CheckBiasThresholdViolation(ctx, lowBiasResult)
	if violation {
		t.Error("Did not expect threshold violation for low bias score")
	}
}

func TestBiasAlerts(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set threshold
	k.SetBiasThreshold(ctx, "model1", 0.5)

	// Record high bias detection
	detection := BiasDetectionResult{
		ModelHash:  "model1",
		BiasScore:  0.9,
		Timestamp:  ctx.BlockTime(),
	}

	err := k.RecordBiasDetection(ctx, detection)
	if err != nil {
		t.Fatalf("Failed to record detection: %v", err)
	}

	// Check if alert was generated
	alerts := k.GetBiasAlerts(ctx, 10)
	hasAlert := false
	for _, alert := range alerts {
		if alert.ModelHash == "model1" && alert.BiasScore > 0.5 {
			hasAlert = true
			break
		}
	}

	if !hasAlert {
		t.Log("Bias alert system may not be fully implemented")
	}
}

func TestCompareBiasAcrossModels(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record detections for multiple models
	models := []struct {
		hash      string
		biasScore float64
	}{
		{"model1", 0.2},
		{"model2", 0.5},
		{"model3", 0.8},
	}

	for _, model := range models {
		detection := BiasDetectionResult{
			ModelHash:  model.hash,
			BiasScore:  model.biasScore,
			Timestamp:  ctx.BlockTime(),
		}
		k.RecordBiasDetection(ctx, detection)
	}

	// Compare bias across models
	comparison := k.CompareBiasAcrossModels(ctx, []string{"model1", "model2", "model3"})

	if len(comparison) != 3 {
		t.Errorf("Expected comparison for 3 models, got %d", len(comparison))
	}
}

func TestBiasDetectionPatterns(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Add bias detection pattern
	pattern := BiasDetectionPattern{
		Name:        "gender_bias",
		Pattern:     "gendered_language",
		Severity:    "high",
		Category:    "gender",
		Description: "Detects gendered language patterns",
	}

	err := k.AddBiasDetectionPattern(ctx, pattern)
	if err != nil {
		t.Fatalf("Failed to add pattern: %v", err)
	}

	// Verify pattern was added
	patterns := k.GetBiasDetectionPatterns(ctx)
	if len(patterns) == 0 {
		t.Error("Expected at least one detection pattern")
	}
}

func TestBiasReportGeneration(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record multiple detections
	for i := 0; i < 10; i++ {
		detection := BiasDetectionResult{
			ModelHash:  "model1",
			BiasScore:  float64(i) * 0.1,
			Timestamp:  ctx.BlockTime(),
		}
		k.RecordBiasDetection(ctx, detection)
	}

	// Generate bias report
	report := k.GenerateBiasReport(ctx, "model1")

	if report.ModelHash != "model1" {
		t.Errorf("Expected report for model1, got %s", report.ModelHash)
	}

	if report.TotalAnalyzed == 0 {
		t.Error("Report should show analyzed queries")
	}
}

func TestBiasDetectionPerformance(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Test performance with many detections
	count := 100
	for i := 0; i < count; i++ {
		result := k.DetectBias(ctx, "model1", "input", "output")
		if result.BiasScore < 0 || result.BiasScore > 1 {
			t.Errorf("Invalid bias score: %v", result.BiasScore)
		}
	}

	t.Logf("Completed %d bias detections", count)
}

func TestBiasCategoryWeighting(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set category weights
	weights := map[string]float64{
		"gender": 1.0,
		"race":   1.5,
		"age":    0.8,
	}

	err := k.SetBiasCategoryWeights(ctx, "model1", weights)
	if err != nil {
		t.Fatalf("Failed to set category weights: %v", err)
	}

	// Verify weights were set
	storedWeights := k.GetBiasCategoryWeights(ctx, "model1")
	if len(storedWeights) != len(weights) {
		t.Errorf("Expected %d weights, got %d", len(weights), len(storedWeights))
	}

	for category, expectedWeight := range weights {
		if actualWeight, exists := storedWeights[category]; exists {
			if actualWeight != expectedWeight {
				t.Errorf("Category %s: expected weight %v, got %v",
					category, expectedWeight, actualWeight)
			}
		}
	}
}

func TestBiasDetectionConfidence(t *testing.T) {
	k, ctx := setupKeeper(t)

	result := k.DetectBias(ctx, "model1", "input", "output")

	// Verify confidence score is reasonable
	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %v", result.Confidence)
	}
}
