package main

import "testing"

func TestExtractUsageFromJSONUsageWrapper(t *testing.T) {
	input, output := extractUsageFromJSON(`{"usage":{"prompt_tokens":161,"completion_tokens":3,"total_tokens":164}}`)
	if input != 161 || output != 3 {
		t.Fatalf("extractUsageFromJSON usage wrapper = (%d, %d), want (161, 3)", input, output)
	}
}

func TestExtractUsageFromJSONFlatTokens(t *testing.T) {
	input, output := extractUsageFromJSON(`{"prompt_tokens":161,"completion_tokens":3,"total_tokens":164}`)
	if input != 161 || output != 3 {
		t.Fatalf("extractUsageFromJSON flat tokens = (%d, %d), want (161, 3)", input, output)
	}
}

func TestExtractTokenUsageStreamingFlatTokens(t *testing.T) {
	resp := "event: message\n" +
		`data: {"type":"message_start","prompt_tokens":161}` + "\n\n" +
		`data: {"type":"message_delta","completion_tokens":3,"total_tokens":164}` + "\n\n"
	input, output := extractTokenUsage(resp)
	if input != 161 || output != 3 {
		t.Fatalf("extractTokenUsage stream flat tokens = (%d, %d), want (161, 3)", input, output)
	}
}
