package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/akulkarni/0perator/internal/runtime"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ExecuteInput defines the input for the execute tool
type ExecuteInput struct {
	Operation string                 `json:"operation" jsonschema:"Operation to execute: run_command, read_file, create_file, edit_file, start_process, stop_process, get_logs, or list_processes"`
	Params    map[string]interface{} `json:"params" jsonschema:"Parameters for the operation"`
}

// ExecuteOutput defines the output for the execute tool
type ExecuteOutput struct {
	Operation string                 `json:"operation"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func (s *Server) handleExecute(ctx context.Context, req *mcp.CallToolRequest, input ExecuteInput) (*mcp.CallToolResult, ExecuteOutput, error) {
	var result ExecuteOutput
	var err error

	switch input.Operation {
	case "run_command":
		result, err = s.executeRunCommand(ctx, input.Params)
	case "read_file":
		result, err = s.executeReadFile(ctx, input.Params)
	case "create_file":
		result, err = s.executeCreateFile(ctx, input.Params)
	case "edit_file":
		result, err = s.executeEditFile(ctx, input.Params)
	case "start_process":
		result, err = s.executeStartProcess(ctx, input.Params)
	case "stop_process":
		result, err = s.executeStopProcess(ctx, input.Params)
	case "get_logs":
		result, err = s.executeGetLogs(ctx, input.Params)
	case "list_processes":
		result, err = s.executeListProcesses(ctx, input.Params)
	default:
		return nil, ExecuteOutput{}, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	if err != nil {
		return nil, ExecuteOutput{}, err
	}

	return nil, result, nil
}

// ===== Primitive 1: run_command =====

func (s *Server) executeRunCommand(ctx context.Context, params map[string]interface{}) (ExecuteOutput, error) {
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return ExecuteOutput{}, fmt.Errorf("missing required parameter: command")
	}

	// Get optional working directory
	workDir, _ := params["cwd"].(string)
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	// Execute command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ExecuteOutput{
			Operation: "run_command",
			Success:   false,
			Message:   fmt.Sprintf("Command failed: %s\n\nOutput:\n%s", err.Error(), string(output)),
			Data: map[string]interface{}{
				"command": command,
				"output":  string(output),
				"error":   err.Error(),
			},
		}, nil
	}

	return ExecuteOutput{
		Operation: "run_command",
		Success:   true,
		Message:   fmt.Sprintf("Command executed successfully\n\nOutput:\n%s", string(output)),
		Data: map[string]interface{}{
			"command": command,
			"output":  string(output),
		},
	}, nil
}

// ===== Primitive 2: read_file =====

func (s *Server) executeReadFile(ctx context.Context, params map[string]interface{}) (ExecuteOutput, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ExecuteOutput{}, fmt.Errorf("missing required parameter: path")
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return ExecuteOutput{
			Operation: "read_file",
			Success:   false,
			Message:   fmt.Sprintf("Failed to read file: %s", err.Error()),
			Data: map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			},
		}, nil
	}

	return ExecuteOutput{
		Operation: "read_file",
		Success:   true,
		Message:   fmt.Sprintf("Read file: %s (%d bytes)", path, len(content)),
		Data: map[string]interface{}{
			"path":    path,
			"content": string(content),
			"size":    len(content),
		},
	}, nil
}

// ===== Primitive 3: create_file =====

func (s *Server) executeCreateFile(ctx context.Context, params map[string]interface{}) (ExecuteOutput, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ExecuteOutput{}, fmt.Errorf("missing required parameter: path")
	}

	content, ok := params["content"].(string)
	if !ok {
		content = ""
	}

	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ExecuteOutput{
			Operation: "create_file",
			Success:   false,
			Message:   fmt.Sprintf("Failed to create directories: %s", err.Error()),
			Data: map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			},
		}, nil
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ExecuteOutput{
			Operation: "create_file",
			Success:   false,
			Message:   fmt.Sprintf("Failed to create file: %s", err.Error()),
			Data: map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			},
		}, nil
	}

	return ExecuteOutput{
		Operation: "create_file",
		Success:   true,
		Message:   fmt.Sprintf("Created file: %s (%d bytes)", path, len(content)),
		Data: map[string]interface{}{
			"path": path,
			"size": len(content),
		},
	}, nil
}

// ===== Primitive 4: edit_file =====

func (s *Server) executeEditFile(ctx context.Context, params map[string]interface{}) (ExecuteOutput, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ExecuteOutput{}, fmt.Errorf("missing required parameter: path")
	}

	oldStr, ok := params["old_string"].(string)
	if !ok || oldStr == "" {
		return ExecuteOutput{}, fmt.Errorf("missing required parameter: old_string")
	}

	newStr, ok := params["new_string"].(string)
	if !ok {
		newStr = ""
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return ExecuteOutput{
			Operation: "edit_file",
			Success:   false,
			Message:   fmt.Sprintf("Failed to read file: %s", err.Error()),
			Data: map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			},
		}, nil
	}

	// Replace string
	oldContent := string(content)
	if !strings.Contains(oldContent, oldStr) {
		return ExecuteOutput{
			Operation: "edit_file",
			Success:   false,
			Message:   fmt.Sprintf("String not found in file: %s", oldStr),
			Data: map[string]interface{}{
				"path":       path,
				"old_string": oldStr,
			},
		}, nil
	}

	newContent := strings.Replace(oldContent, oldStr, newStr, 1)

	// Write file
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return ExecuteOutput{
			Operation: "edit_file",
			Success:   false,
			Message:   fmt.Sprintf("Failed to write file: %s", err.Error()),
			Data: map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			},
		}, nil
	}

	return ExecuteOutput{
		Operation: "edit_file",
		Success:   true,
		Message:   fmt.Sprintf("Edited file: %s", path),
		Data: map[string]interface{}{
			"path": path,
		},
	}, nil
}

// ===== Primitive 5: start_process =====

func (s *Server) executeStartProcess(ctx context.Context, params map[string]interface{}) (ExecuteOutput, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ExecuteOutput{}, fmt.Errorf("missing required parameter: path")
	}

	// Get optional port
	port := 0
	if portFloat, ok := params["port"].(float64); ok {
		port = int(portFloat)
	}

	// Start process
	process, err := s.processManager.Deploy(ctx, path, port)
	if err != nil {
		return ExecuteOutput{
			Operation: "start_process",
			Success:   false,
			Message:   fmt.Sprintf("Failed to start process: %s", err.Error()),
			Data: map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			},
		}, nil
	}

	// Get recent logs
	logs := runtime.TailLogs(process.LogFile.Name(), 10)

	return ExecuteOutput{
		Operation: "start_process",
		Success:   true,
		Message:   fmt.Sprintf("Started process: %s\nPort: %d\nURL: http://localhost:%d\n\nRecent logs:\n%s", process.ID, process.Port, process.Port, logs),
		Data: map[string]interface{}{
			"process_id": process.ID,
			"path":       process.Path,
			"port":       process.Port,
			"url":        fmt.Sprintf("http://localhost:%d", process.Port),
			"log_file":   process.LogFile.Name(),
		},
	}, nil
}

// ===== Primitive 6: stop_process =====

func (s *Server) executeStopProcess(ctx context.Context, params map[string]interface{}) (ExecuteOutput, error) {
	processID, ok := params["process_id"].(string)
	if !ok || processID == "" {
		return ExecuteOutput{}, fmt.Errorf("missing required parameter: process_id")
	}

	// Stop process
	if err := s.processManager.Stop(processID); err != nil {
		return ExecuteOutput{
			Operation: "stop_process",
			Success:   false,
			Message:   fmt.Sprintf("Failed to stop process: %s", err.Error()),
			Data: map[string]interface{}{
				"process_id": processID,
				"error":      err.Error(),
			},
		}, nil
	}

	return ExecuteOutput{
		Operation: "stop_process",
		Success:   true,
		Message:   fmt.Sprintf("Stopped process: %s", processID),
		Data: map[string]interface{}{
			"process_id": processID,
		},
	}, nil
}

// ===== Primitive 7: get_logs =====

func (s *Server) executeGetLogs(ctx context.Context, params map[string]interface{}) (ExecuteOutput, error) {
	processID, ok := params["process_id"].(string)
	if !ok || processID == "" {
		return ExecuteOutput{}, fmt.Errorf("missing required parameter: process_id")
	}

	// Get optional lines count
	lines := 50
	if linesFloat, ok := params["lines"].(float64); ok && linesFloat > 0 {
		lines = int(linesFloat)
	}

	// Get process
	process, exists := s.processManager.GetProcess(processID)
	if !exists {
		return ExecuteOutput{
			Operation: "get_logs",
			Success:   false,
			Message:   fmt.Sprintf("Process not found: %s", processID),
			Data: map[string]interface{}{
				"process_id": processID,
			},
		}, nil
	}

	// Get logs
	logs := runtime.TailLogs(process.LogFile.Name(), lines)

	return ExecuteOutput{
		Operation: "get_logs",
		Success:   true,
		Message:   fmt.Sprintf("Logs for process %s:\n\n%s", processID, logs),
		Data: map[string]interface{}{
			"process_id": processID,
			"logs":       logs,
			"lines":      lines,
		},
	}, nil
}

// ===== Primitive 8: list_processes =====

func (s *Server) executeListProcesses(ctx context.Context, params map[string]interface{}) (ExecuteOutput, error) {
	processes := s.processManager.ListProcesses()

	if len(processes) == 0 {
		return ExecuteOutput{
			Operation: "list_processes",
			Success:   true,
			Message:   "No processes currently running",
			Data: map[string]interface{}{
				"processes": []interface{}{},
			},
		}, nil
	}

	// Build process list
	processInfos := make([]map[string]interface{}, 0, len(processes))
	message := "Running processes:\n\n"

	for i, proc := range processes {
		message += fmt.Sprintf("%d. Process ID: %s\n", i+1, proc.ID)
		message += fmt.Sprintf("   Path:     %s\n", proc.Path)
		message += fmt.Sprintf("   Port:     %d\n", proc.Port)
		message += fmt.Sprintf("   URL:      http://localhost:%d\n", proc.Port)
		message += fmt.Sprintf("   Log File: %s\n\n", proc.LogFile.Name())

		processInfos = append(processInfos, map[string]interface{}{
			"process_id": proc.ID,
			"path":       proc.Path,
			"port":       proc.Port,
			"url":        fmt.Sprintf("http://localhost:%d", proc.Port),
			"log_file":   proc.LogFile.Name(),
		})
	}

	return ExecuteOutput{
		Operation: "list_processes",
		Success:   true,
		Message:   message,
		Data: map[string]interface{}{
			"processes": processInfos,
		},
	}, nil
}
