package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/antnet1094/plf/pkg/parser"
	"github.com/antnet1094/plf/pkg/renderer"
	"github.com/antnet1094/plf/pkg/types"
)

type SessionLog struct {
	SessionID      string    `json:"session_id"`
	AgentName      string    `json:"agent_name"`
	StartTime      time.Time `json:"start_time"`
	CurrentTask    string    `json:"current_task"`
	Thoughts       []Thought `json:"thoughts"`
	Actions        []Action  `json:"actions"`
	Progress       int       `json:"progress_percent"`
	TotalSteps     int       `json:"total_steps"`
	CompletedSteps int       `json:"completed_steps"`
}

type Thought struct {
	Timestamp   time.Time `json:"timestamp"`
	Content     string    `json:"content"`
	Confidence  string    `json:"confidence"`
	Updated     bool      `json:"updated,omitempty"`
	PrevThought string    `json:"prev_thought,omitempty"`
}

type Action struct {
	Timestamp     time.Time `json:"timestamp"`
	Type          string    `json:"type"` // command, read, write, search, test
	Command       string    `json:"command,omitempty"`
	Description   string    `json:"description"`
	Result        string    `json:"result"` // success, partial, failure
	Output        string    `json:"output,omitempty"`
	FilesCreated  []string  `json:"files_created,omitempty"`
	FilesModified []string  `json:"files_modified,omitempty"`
	Error         string    `json:"error,omitempty"`
}

func main() {
	mode := flag.String("mode", "generate", "Mode: generate, log-thought, log-action, export")
	sessionID := flag.String("session", "", "Session ID")
	agentName := flag.String("agent", "gemini-cli", "Agent name")
	task := flag.String("task", "", "Task description")
	template := flag.String("template", "examples/agent_session_logger.plf", "PLF template")
	output := flag.String("output", "", "Output file (default: session_<id>.plf)")

	flag.Parse()

	switch *mode {
	case "generate":
		generateSession(*template, *task, *sessionID, *agentName, *output)
	case "log-thought":
		logThought(*sessionID, strings.Join(flag.Args(), " "))
	case "log-action":
		logAction(*sessionID, strings.Join(flag.Args(), " "))
	case "export":
		exportSession(*sessionID)
	case "status":
		showStatus(*sessionID)
	case "training":
		startTraining(*sessionID)
	default:
		fmt.Println("Unknown mode:", *mode)
		os.Exit(1)
	}
}

func generateSession(templatePath, task, sessionID, agentName, outputPath string) {
	if sessionID == "" {
		sessionID = fmt.Sprintf("session_%d", time.Now().Unix())
	}
	if outputPath == "" {
		outputPath = fmt.Sprintf("session_%s.plf", sessionID)
	}

	vars := map[string]string{
		"session_id":             sessionID,
		"agent_name":             agentName,
		"task_description":       task,
		"timestamp_start":        time.Now().Format(time.RFC3339),
		"task_type":              detectTaskType(task),
		"priority":               "normal",
		"constraints":            "None specified",
		"expected_output":        "Documented solution",
		"current_thought":        "Initializing session...",
		"decision_log":           "N/A - Session just started",
		"actions_log":            "N/A - No actions yet",
		"command_log":            "N/A",
		"verification_steps":     "Pending",
		"test_results":           "Pending",
		"error_log":              "None",
		"warnings":               "None",
		"percentage":             "0",
		"status":                 "in_progress",
		"completed_steps":        "None",
		"current_step":           "Initialization",
		"remaining_steps":        "To be determined",
		"new_knowledge":          "N/A",
		"discovered_items":       "N/A",
		"old_assumption":         "N/A",
		"new_assumption":         "N/A",
		"solution_summary":       "Pending",
		"key_points":             "Pending",
		"implementation_details": "Pending",
		"how_to_use":             "Pending",
		"limitations":            "Pending",
		"suggested_next_steps":   "Pending",
		"duration":               "0",
		"count":                  "0",
		"estimate":               "0",
		"cost":                   "0",
		"revision_history":       "v1.0 - Session created",
		"confidence_level":       "MEDIUM",
		"blockers_if_any":        "None",
		"total_thoughts":         "0",
		"total_actions":          "0",
		"total_errors":           "0",
	}

	doc, err := parser.ParseFile(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars: vars,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, []byte(result.Full), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	saveSessionMeta(sessionID, agentName, task, outputPath)

	fmt.Printf("✓ Session generated: %s\n", outputPath)
	fmt.Printf("  Session ID: %s\n", sessionID)
	fmt.Printf("  Task: %s\n", task)
	fmt.Println("\n📝 Use these commands to log during the session:")
	fmt.Printf("   plf-agent log-thought --session %s \"your thought here\"\n", sessionID)
	fmt.Printf("   plf-agent log-action --session %s \"command or action\"\n", sessionID)
	fmt.Printf("   plf-agent status --session %s\n", sessionID)
}

func logThought(sessionID, thought string) {
	logEntry(sessionID, "thought", thought)
	fmt.Printf("✓ Thought logged to session %s\n", sessionID)
}

func logAction(sessionID, action string) {
	logEntry(sessionID, "action", action)
	fmt.Printf("✓ Action logged to session %s\n", sessionID)
}

func logEntry(sessionID, entryType, content string) {
	metaPath := fmt.Sprintf(".plf_sessions/%s.meta", sessionID)

	var log SessionLog
	if data, err := os.ReadFile(metaPath); err == nil {
		json.Unmarshal(data, &log)
	}

	entry := Thought{
		Timestamp:  time.Now(),
		Content:    content,
		Confidence: "MEDIUM",
	}

	if entryType == "thought" {
		log.Thoughts = append(log.Thoughts, entry)
	} else {
		log.Actions = append(log.Actions, Action{
			Timestamp:   time.Now(),
			Type:        "command",
			Command:     content,
			Description: content,
			Result:      "success",
		})
	}

	data, _ := json.MarshalIndent(log, "", "  ")
	os.WriteFile(metaPath, data, 0644)
}

func showStatus(sessionID string) {
	metaPath := fmt.Sprintf(".plf_sessions/%s.meta", sessionID)

	data, err := os.ReadFile(metaPath)
	if err != nil {
		fmt.Printf("Session %s not found\n", sessionID)
		os.Exit(1)
	}

	var log SessionLog
	json.Unmarshal(data, &log)

	duration := time.Since(log.StartTime)

	fmt.Printf("\n📊 Session Status: %s\n\n", sessionID)
	fmt.Printf("  Agent:     %s\n", log.AgentName)
	fmt.Printf("  Task:      %s\n", log.CurrentTask)
	fmt.Printf("  Duration:  %s\n", duration.Round(time.Second))
	fmt.Printf("  Progress:  %d%% (%d/%d steps)\n", log.Progress, log.CompletedSteps, log.TotalSteps)
	fmt.Printf("  Thoughts:  %d logged\n", len(log.Thoughts))
	fmt.Printf("  Actions:   %d logged\n", len(log.Actions))
	fmt.Println()
}

func exportSession(sessionID string) {
	metaPath := fmt.Sprintf(".plf_sessions/%s.meta", sessionID)

	data, err := os.ReadFile(metaPath)
	if err != nil {
		fmt.Printf("Session %s not found\n", sessionID)
		os.Exit(1)
	}

	var log SessionLog
	json.Unmarshal(data, &log)

	output := generatePLFFromSession(&log)
	
	os.MkdirAll("secciones", 0755)
	plfPath := fmt.Sprintf("secciones/session_%s_exported.plf", sessionID)
	os.WriteFile(plfPath, []byte(output), 0644)

	fmt.Printf("✓ Session exported to: %s\n", plfPath)
}

func generatePLFFromSession(log *SessionLog) string {
	var sb strings.Builder

	sb.WriteString("# AGENT SESSION EXPORT\n")
	sb.WriteString("# Generated by PLF Agent Session Logger\n\n")

	sb.WriteString("@meta\n")
	sb.WriteString(fmt.Sprintf("  session_id: %s\n", log.SessionID))
	sb.WriteString(fmt.Sprintf("  agent_name: %s\n", log.AgentName))
	sb.WriteString(fmt.Sprintf("  start_time: %s\n", log.StartTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("  end_time: %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("  duration: %s\n", time.Since(log.StartTime).Round(time.Second)))
	sb.WriteString("\n")

	sb.WriteString("@thoughts\n")
	sb.WriteString("# INNER REASONING LOG\n\n")
	for i, t := range log.Thoughts {
		sb.WriteString(fmt.Sprintf("## Thought #%d\n", i+1))
		sb.WriteString(fmt.Sprintf("- timestamp: %s\n", t.Timestamp.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- content: %s\n", t.Content))
		sb.WriteString(fmt.Sprintf("- confidence: %s\n\n", t.Confidence))
	}

	sb.WriteString("@actions\n")
	sb.WriteString("# COMMAND AND ACTION LOG\n\n")
	for i, a := range log.Actions {
		sb.WriteString(fmt.Sprintf("## Action #%d\n", i+1))
		sb.WriteString(fmt.Sprintf("- timestamp: %s\n", a.Timestamp.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- type: %s\n", a.Type))
		if a.Command != "" {
			sb.WriteString(fmt.Sprintf("- command: `%s`\n", a.Command))
		}
		sb.WriteString(fmt.Sprintf("- description: %s\n", a.Description))
		sb.WriteString(fmt.Sprintf("- result: %s\n", a.Result))
		if a.Error != "" {
			sb.WriteString(fmt.Sprintf("- error: %s\n", a.Error))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("@progress\n")
	sb.WriteString(fmt.Sprintf("PROGRESS: %d%%\n", log.Progress))
	sb.WriteString(fmt.Sprintf("COMPLETED_STEPS: %d/%d\n", log.CompletedSteps, log.TotalSteps))
	sb.WriteString("\n")

	return sb.String()
}

func saveSessionMeta(sessionID, agentName, task, outputPath string) {
	os.MkdirAll(".plf_sessions", 0755)

	log := SessionLog{
		SessionID:      sessionID,
		AgentName:      agentName,
		CurrentTask:    task,
		StartTime:      time.Now(),
		Thoughts:       []Thought{},
		Actions:        []Action{},
		Progress:       0,
		TotalSteps:     0,
		CompletedSteps: 0,
	}

	data, _ := json.MarshalIndent(log, "", "  ")
	metaPath := fmt.Sprintf(".plf_sessions/%s.meta", sessionID)
	os.WriteFile(metaPath, data, 0644)
}

func detectTaskType(task string) string {
	taskLower := strings.ToLower(task)
	if strings.Contains(taskLower, "bug") || strings.Contains(taskLower, "fix") {
		return "bug_fix"
	}
	if strings.Contains(taskLower, "test") {
		return "testing"
	}
	if strings.Contains(taskLower, "doc") || strings.Contains(taskLower, "readme") {
		return "documentation"
	}
	if strings.Contains(taskLower, "refactor") {
		return "refactoring"
	}
	if strings.Contains(taskLower, "feature") || strings.Contains(taskLower, "add") {
		return "feature_development"
	}
	return "general"
}
