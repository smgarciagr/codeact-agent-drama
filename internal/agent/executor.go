package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// sanitizeCode removes markdown code fences and other artifacts the LLM may add
func sanitizeCode(code string) string {
	code = strings.TrimSpace(code)
	// Remove markdown code fences like ```go or ```
	if strings.HasPrefix(code, "```") {
		// Remove first line (```go or ```)
		if idx := strings.Index(code, "\n"); idx != -1 {
			code = code[idx+1:]
		}
	}
	// Remove trailing ```
	if strings.HasSuffix(code, "```") {
		code = strings.TrimSuffix(code, "```")
	}
	return strings.TrimSpace(code)
}

// ExecuteCode saves the code to a temporary file and runs it with a timeout
func ExecuteCode(code string) (string, error) {
	code = sanitizeCode(code)

	tmpFile := "temp_agent_task.go"
	err := os.WriteFile(tmpFile, []byte(code), 0644)
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %v", err)
	}
	defer os.Remove(tmpFile) // Clean up the file after use

	// Execute with a 30-second timeout to prevent infinite loops
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", tmpFile)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("execution timed out after 30 seconds")
	}

	if err != nil {
		return string(output), fmt.Errorf("error executing code: %v", err)
	}
	return string(output), nil
}
