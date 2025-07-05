package taskcp_test

import (
	"fmt"

	"github.com/gopatchy/taskcp"
	"github.com/mark3labs/mcp-go/server"
)

func ExampleRegisterMCPTools() {
	service := taskcp.New()

	project := service.AddProject()
	fmt.Printf("Created project: %s\n", project.ID)

	task1 := project.InsertTaskBefore("", "Compile the code", func(task *taskcp.Task) {
		fmt.Printf("Task %s completed with state: %s\n", task.ID, task.State)
	})
	
	task2 := project.InsertTaskBefore("", "Run tests", func(task *taskcp.Task) {
		fmt.Printf("Task %s completed with state: %s\n", task.ID, task.State)
	})
	
	task1.NextTaskID = task2.ID
	project.NextTaskID = task1.ID

	mcpServer := server.NewMCPServer(
		"TaskCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	err := taskcp.RegisterMCPTools(mcpServer, service)
	if err != nil {
		fmt.Printf("Failed to register tools: %v\n", err)
		return
	}

	fmt.Println("MCP tools registered successfully")
	
}