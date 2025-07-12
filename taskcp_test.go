package taskcp_test

import (
	"fmt"
	"testing"

	"github.com/gopatchy/taskcp"
	"github.com/stretchr/testify/require"
)

func TestPlaceholderExpansion(t *testing.T) {
	service := taskcp.New("my_service")
	p := service.AddProject()

	p.AddLastTask().
		WithTitle("Please complete this task.").
		WithInstructions("{SUCCESS_PROMPT}").
		Then(func(task *taskcp.Task) error {
			return nil
		})

	task, err := p.PopNextTask()
	require.NoError(t, err)
	require.NotNil(t, task)
	require.Contains(t, task.Instructions, "my_service.set_task_success")
	require.NotContains(t, task.Instructions, "{SUCCESS_PROMPT}")

	p.AddLastTask().
		WithTitle("Try this risky operation.").
		WithInstructions("{FAILURE_PROMPT}").
		Then(func(task *taskcp.Task) error {
			return nil
		})

	task, err = p.PopNextTask()
	require.NoError(t, err)
	require.NotNil(t, task)
	require.Contains(t, task.Instructions, "my_service.set_task_failure")
	require.NotContains(t, task.Instructions, "{FAILURE_PROMPT}")
}

func TestTaskFlow(t *testing.T) {
	service := taskcp.New("test_service")
	p := service.AddProject()

	var completed []string

	p.AddLastTask().
		WithTitle("First task").
		Then(func(task *taskcp.Task) error {
			completed = append(completed, task.Title)
			return nil
		})

	p.AddLastTask().
		WithTitle("Second task").
		Then(func(task *taskcp.Task) error {
			completed = append(completed, task.Title)
			return nil
		})

	task1, err := p.PopNextTask()
	require.NoError(t, err)
	require.NotNil(t, task1)
	require.Equal(t, "First task", task1.Title)

	task2, err := task1.SetSuccess("Task 1 done", "")
	require.NoError(t, err)
	require.NotNil(t, task2)
	require.Equal(t, "Second task", task2.Title)

	task3, err := task2.SetFailure("Task 2 failed", "Error details")
	require.NoError(t, err)
	require.Nil(t, task3)

	require.Equal(t, []string{"First task", "Second task"}, completed)
	require.Equal(t, "Task 1 done", task1.Result)
	require.Equal(t, "Task 2 failed", task2.Error)
}

func TestCallbackError(t *testing.T) {
	service := taskcp.New("test_service")
	p := service.AddProject()

	expectedErr := fmt.Errorf("callback error")

	p.AddLastTask().
		WithTitle("Task with error callback").
		WithInstructions("This is a test task.").
		WithData("key", "value").
		Then(func(task *taskcp.Task) error {
			return expectedErr
		})

	task, err := p.PopNextTask()
	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, "Task with error callback", task.Title)

	_, err = task.SetSuccess("Result", "")
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}
