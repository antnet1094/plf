package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/antnet1094/plf/pkg/parser"
	"github.com/antnet1094/plf/pkg/renderer"
	"github.com/antnet1094/plf/pkg/types"
)

type TestCase struct {
	Name            string            `json:"name"`
	Vars            map[string]string `json:"vars"`
	ExpectedInclude string            `json:"expected_include"`
	ExpectedExclude string            `json:"expected_exclude"`
}

type TestSuite struct {
	TargetEndpoint string     `json:"target_endpoint"` // e.g., "http://127.0.0.1:8081/v1/chat/completions"
	Model          string     `json:"model"`
	Tests          []TestCase `json:"tests"`
}

type EvalResult struct {
	CaseName string
	Passed   bool
	Latency  time.Duration
	Error    error
	Response string
}

func RunEval(plfPath string, suitePath string) ([]EvalResult, error) {
	doc, err := parser.ParseFile(plfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", plfPath, err)
	}

	b, err := loadFile(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test suite %s: %w", suitePath, err)
	}

	var suite TestSuite
	if err := json.Unmarshal(b, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse test suite JSON: %w", err)
	}

	if suite.TargetEndpoint == "" {
		suite.TargetEndpoint = "http://127.0.0.1:8081/v1/chat/completions"
	}

	var results []EvalResult

	for _, tc := range suite.Tests {
		res := runTestCase(doc, suite.TargetEndpoint, suite.Model, tc)
		results = append(results, res)
	}

	return results, nil
}

func loadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func runTestCase(doc *types.Document, endpoint, model string, tc TestCase) EvalResult {
	start := time.Now()
	
	rendered, err := renderer.Render(doc, types.RenderOptions{
		Vars: tc.Vars,
		Format: types.FormatRaw,
	})
	
	if err != nil {
		return EvalResult{CaseName: tc.Name, Passed: false, Error: err}
	}

	// Assuming a standard OpenAI API compatible payload
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": rendered.System},
			{"role": "user", "content": rendered.User},
		},
		"temperature": 0.0,
	}

	jsonBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonBytes))
	
	if err != nil {
		return EvalResult{CaseName: tc.Name, Passed: false, Error: err, Latency: time.Since(start)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return EvalResult{CaseName: tc.Name, Passed: false, Error: fmt.Errorf("HTTP %d", resp.StatusCode), Latency: time.Since(start)}
	}

	body, _ := io.ReadAll(resp.Body)
	
	// Quick parse assuming OpenAI standard `choices[0].message.content`
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	
	if err := json.Unmarshal(body, &payload); err != nil {
		return EvalResult{CaseName: tc.Name, Passed: false, Error: fmt.Errorf("failed to parse LLM response: %v", err), Latency: time.Since(start)}
	}

	content := ""
	if len(payload.Choices) > 0 {
		content = payload.Choices[0].Message.Content
	} else {
		return EvalResult{CaseName: tc.Name, Passed: false, Error: fmt.Errorf("no choices returned"), Latency: time.Since(start)}
	}

	passed := true
	if tc.ExpectedInclude != "" && !strings.Contains(strings.ToLower(content), strings.ToLower(tc.ExpectedInclude)) {
		passed = false
	}
	if tc.ExpectedExclude != "" && strings.Contains(strings.ToLower(content), strings.ToLower(tc.ExpectedExclude)) {
		passed = false
	}

	return EvalResult{
		CaseName: tc.Name,
		Passed:   passed,
		Latency:  time.Since(start),
		Response: content,
	}
}
