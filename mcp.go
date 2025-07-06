package taskcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type setTaskSuccessArgs struct {
	ProjectID string `json:"project_id"`
	TaskID    string `json:"task_id"`
	Result    string `json:"result"`
	Notes     string `json:"notes,omitempty"`
}

type setTaskFailureArgs struct {
	ProjectID string `json:"project_id"`
	TaskID    string `json:"task_id"`
	Error     string `json:"error"`
	Notes     string `json:"notes,omitempty"`
}

type taskResponse struct {
	TaskID   string `json:"task_id"`
	Message  string `json:"message"`
	NextTask *Task  `json:"next_task,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type ServiceHandlerFunc[TArgs any, TResponse any] func(s *Service, ctx context.Context, args TArgs) (*TResponse, error)

func wrapServiceHandler[TArgs any, TResponse any](s *Service, handler ServiceHandlerFunc[TArgs, TResponse]) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args TArgs
		if err := request.BindArguments(&args); err != nil {
			errorJSON, _ := json.Marshal(errorResponse{Error: err.Error()})
			return mcp.NewToolResultText(string(errorJSON)), nil
		}

		response, err := handler(s, ctx, args)
		if err != nil {
			errorJSON, _ := json.Marshal(errorResponse{Error: err.Error()})
			return mcp.NewToolResultText(string(errorJSON)), nil
		}

		resultJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			errorJSON, _ := json.Marshal(errorResponse{Error: err.Error()})
			return mcp.NewToolResultText(string(errorJSON)), nil
		}

		return mcp.NewToolResultText(string(resultJSON)), nil
	}
}

func (s *Service) RegisterMCPTools(mcpServer *server.MCPServer) error {
	mcpServer.AddTool(
		mcp.NewTool(
			"set_task_success",
			mcp.WithDescription("Mark a task as successfully completed"),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("The project ID"),
			),
			mcp.WithString("task_id",
				mcp.Required(),
				mcp.Description("The task ID to mark as successful"),
			),
			mcp.WithString("result",
				mcp.Required(),
				mcp.Description("The result of the task execution"),
			),
			mcp.WithString("notes",
				mcp.Description("Additional notes about the task completion"),
			),
		),
		wrapServiceHandler(s, handleSetTaskSuccess),
	)

	mcpServer.AddTool(
		mcp.NewTool(
			"set_task_failure",
			mcp.WithDescription("Mark a task as failed"),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("The project ID"),
			),
			mcp.WithString("task_id",
				mcp.Required(),
				mcp.Description("The task ID to mark as failed"),
			),
			mcp.WithString("error",
				mcp.Required(),
				mcp.Description("The error message describing why the task failed"),
			),
			mcp.WithString("notes",
				mcp.Description("Additional notes about the task failure"),
			),
		),
		wrapServiceHandler(s, handleSetTaskFailure),
	)

	return nil
}

func handleSetTaskSuccess(s *Service, ctx context.Context, args setTaskSuccessArgs) (*taskResponse, error) {
	project, err := s.GetProject(args.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	nextTask, err := project.SetTaskSuccess(args.TaskID, args.Result, args.Notes)
	if err != nil {
		return nil, fmt.Errorf("completion callback error: %w", err)
	}

	response := &taskResponse{
		TaskID:   args.TaskID,
		Message:  fmt.Sprintf("Task %s marked as successful", args.TaskID),
		NextTask: nextTask,
	}

	return response, nil
}

func handleSetTaskFailure(s *Service, ctx context.Context, args setTaskFailureArgs) (*taskResponse, error) {
	project, err := s.GetProject(args.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	nextTask, err := project.SetTaskFailure(args.TaskID, args.Error, args.Notes)
	if err != nil {
		return nil, fmt.Errorf("completion callback error: %w", err)
	}

	response := &taskResponse{
		TaskID:   args.TaskID,
		Message:  fmt.Sprintf("Task %s marked as failed", args.TaskID),
		NextTask: nextTask,
	}

	return response, nil
}