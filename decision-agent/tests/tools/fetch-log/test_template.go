package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LogDetail represents a single log file entry
type LogDetail struct {
	LogName    string `json:"log_name"`
	LogContent string `json:"log_content"`
}

// LogEntry represents a single log entry for a service
type LogEntry struct {
	SvcName       string      `json:"svc_name"`
	Pod           string      `json:"pod"`
	FetchTime     string      `json:"fetch_time"`
	FetchLogLines int         `json:"fecth_log_lines"`
	LogDetail     []LogDetail `json:"log_detail"`
}

// LogResult represents the collection of log entries
type LogResult []LogEntry

func main() {
	// Create a test log result
	testResult := LogResult{
		{
			SvcName:       "agent-app",
			Pod:           "agent-app-test-pod",
			FetchTime:     "2026-01-12 12:00:00",
			FetchLogLines: 10,
			LogDetail: []LogDetail{
				{
					LogName:    "agent-app",
					LogContent: "Test log content for agent-app",
				},
			},
		},
	}

	// Use current working directory instead of executable directory
	execDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Working directory: %s\n", execDir)

	// Read template file
	templatePath := filepath.Join(execDir, "prompts", "log_analysis.txt")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("❌ Error reading template: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Template file loaded: %s\n", templatePath)

	// Read Dolphin documentation
	dolphinDocPath := filepath.Join(execDir, "ref", "dolphin_frendly_program.txt")
	dolphinDoc, err := os.ReadFile(dolphinDocPath)
	if err != nil {
		fmt.Printf("❌ Error reading dolphin doc: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Dolphin doc loaded: %s (%d bytes)\n", dolphinDocPath, len(dolphinDoc))

	// Read knowledge base
	knowledgeBasePath := filepath.Join(execDir, "ref", "agent_knowledge_items.json")
	knowledgeBase, err := os.ReadFile(knowledgeBasePath)
	if err != nil {
		fmt.Printf("❌ Error reading knowledge base: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Knowledge base loaded: %s (%d bytes)\n", knowledgeBasePath, len(knowledgeBase))

	// Convert log result to JSON
	logJSON, err := json.MarshalIndent(testResult, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling log: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Log result marshaled (%d bytes)\n", len(logJSON))

	// Replace placeholders
	template := string(templateContent)
	template = strings.ReplaceAll(template, "to_fill_error_log", string(logJSON))
	template = strings.ReplaceAll(template, "to_fill_dolphin_doc", string(dolphinDoc))
	template = strings.ReplaceAll(template, "to_fill_knowledge_items", string(knowledgeBase))

	// Check if placeholders were replaced
	if strings.Contains(template, "to_fill_error_log") {
		fmt.Println("❌ Error: to_fill_error_log was not replaced!")
		os.Exit(1)
	}
	if strings.Contains(template, "to_fill_dolphin_doc") {
		fmt.Println("❌ Error: to_fill_dolphin_doc was not replaced!")
		os.Exit(1)
	}
	if strings.Contains(template, "to_fill_knowledge_items") {
		fmt.Println("❌ Error: to_fill_knowledge_items was not replaced!")
		os.Exit(1)
	}

	fmt.Println("\n✅ All placeholders successfully replaced!")
	fmt.Printf("\n📋 Template preview (first 500 chars):\n%s\n", template[:min(500, len(template))])
	fmt.Println("\n✅ Template replacement test PASSED!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
