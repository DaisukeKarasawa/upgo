package llm

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestClient_Generate_Performance measures LLM generation performance
func TestClient_Generate_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// This test requires a running Ollama instance
	// Skip if Ollama is not available
	baseURL := "http://localhost:11434"
	model := "llama3.2"
	timeout := 30
	logger := zap.NewNop()

	client := NewClient(baseURL, model, timeout, logger)

	ctx := context.Background()

	// Check connection first
	if err := client.CheckConnection(ctx); err != nil {
		t.Skipf("Ollama is not available: %v", err)
	}

	// Test prompt
	prompt := "こんにちは、これはパフォーマンステストです。"

	// Measure generation time
	iterations := 3
	var totalDuration time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		result, err := client.Generate(ctx, prompt)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		totalDuration += duration
		t.Logf("Iteration %d: Duration: %v, Result length: %d", i+1, duration, len(result))
	}

	avgDuration := totalDuration / time.Duration(iterations)

	t.Logf("\n=== LLM Performance Metrics ===")
	t.Logf("  Model: %s", model)
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  Average Duration: %v", avgDuration)
	t.Logf("  Average Duration (ms): %.2f", float64(avgDuration.Nanoseconds())/1e6)
	t.Logf("  Generations per second: %.2f", float64(iterations)/totalDuration.Seconds())
}

// BenchmarkClient_Generate benchmarks LLM generation
func BenchmarkClient_Generate(b *testing.B) {
	baseURL := "http://localhost:11434"
	model := "llama3.2"
	timeout := 30
	logger := zap.NewNop()

	client := NewClient(baseURL, model, timeout, logger)

	ctx := context.Background()

	// Check connection first
	if err := client.CheckConnection(ctx); err != nil {
		b.Skipf("Ollama is not available: %v", err)
	}

	prompt := "こんにちは、これはベンチマークテストです。"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.Generate(ctx, prompt)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

// TestSummarizer_Performance measures summarizer performance
func TestSummarizer_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	baseURL := "http://localhost:11434"
	model := "llama3.2"
	timeout := 30
	logger := zap.NewNop()

	llmClient := NewClient(baseURL, model, timeout, logger)
	summarizer := NewSummarizer(llmClient, logger)

	ctx := context.Background()

	// Check connection first
	if err := llmClient.CheckConnection(ctx); err != nil {
		t.Skipf("Ollama is not available: %v", err)
	}

	// Test data
	description := "This is a test PR description that needs to be summarized. " +
		"It contains multiple sentences to test the summarization performance. " +
		"The summarizer should be able to handle this text efficiently."

	// Measure SummarizeDescription performance
	start := time.Now()
	result, err := summarizer.SummarizeDescription(ctx, description)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("SummarizeDescription failed: %v", err)
	}

	t.Logf("\n=== Summarizer Performance Metrics ===")
	t.Logf("  Operation: SummarizeDescription")
	t.Logf("  Input length: %d", len(description))
	t.Logf("  Output length: %d", len(result))
	t.Logf("  Duration: %v", duration)
	t.Logf("  Duration (ms): %.2f", float64(duration.Nanoseconds())/1e6)
	t.Logf("  Characters per second: %.2f", float64(len(description))/duration.Seconds())
}
