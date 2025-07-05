package taskcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMCPTools(s *server.MCPServer, service *Service) error {
	s.AddTool(
		mcp.NewTool(
			"SetTaskSuccess",
			mcp.WithDescription("Mark a task as successfully completed"),
			mcp.WithString("projectId",
				mcp.Required(),
				mcp.Description("The project ID"),
			),
			mcp.WithString("taskId",
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
		handleSetTaskSuccess(service),
	)

	s.AddTool(
		mcp.NewTool(
			"SetTaskFailure",
			mcp.WithDescription("Mark a task as failed"),
			mcp.WithString("projectId",
				mcp.Required(),
				mcp.Description("The project ID"),
			),
			mcp.WithString("taskId",
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
		handleSetTaskFailure(service),
	)

	return nil
}

func handleSetTaskSuccess(service *Service) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectId, err := request.RequireString("projectId")
		if err != nil {
			return nil, fmt.Errorf("failed to get projectId: %w", err)
		}

		taskId, err := request.RequireString("taskId")
		if err != nil {
			return nil, fmt.Errorf("failed to get taskId: %w", err)
		}

		result, err := request.RequireString("result")
		if err != nil {
			return nil, fmt.Errorf("failed to get result: %w", err)
		}

		notes := request.GetString("notes", "")

		project, err := service.GetProject(projectId)
		if err != nil {
			return nil, fmt.Errorf("failed to get project: %w", err)
		}

		nextTask := project.SetTaskSuccess(taskId, result, notes)

		message := fmt.Sprintf("Task %s marked as successful", taskId)
		if nextTask != nil {
			message += fmt.Sprintf("\nNext task: %s (ID: %s)", nextTask.Instructions, nextTask.ID)
		}

		return mcp.NewToolResultText(message), nil
	}
}

func handleSetTaskFailure(service *Service) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectId, err := request.RequireString("projectId")
		if err != nil {
			return nil, fmt.Errorf("failed to get projectId: %w", err)
		}

		taskId, err := request.RequireString("taskId")
		if err != nil {
			return nil, fmt.Errorf("failed to get taskId: %w", err)
		}

		errorMsg, err := request.RequireString("error")
		if err != nil {
			return nil, fmt.Errorf("failed to get error: %w", err)
		}

		notes := request.GetString("notes", "")

		project, err := service.GetProject(projectId)
		if err != nil {
			return nil, fmt.Errorf("failed to get project: %w", err)
		}

		nextTask := project.SetTaskFailure(taskId, errorMsg, notes)

		message := fmt.Sprintf("Task %s marked as failed", taskId)
		if nextTask != nil {
			message += fmt.Sprintf("\nNext task: %s (ID: %s)", nextTask.Instructions, nextTask.ID)
		}

		return mcp.NewToolResultText(message), nil
	}
}