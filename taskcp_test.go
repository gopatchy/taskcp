package taskcp_test

import (
	"fmt"
	"testing"

	"github.com/gopatchy/taskcp"
	"github.com/stretchr/testify/require"
)

func TestTaskPrompts(t *testing.T) {
	service := taskcp.New("my_service")
	project := service.AddProject()

	task := project.InsertTaskBefore("", "Write unit tests", "", func(project *taskcp.Project, task *taskcp.Task) error { return nil })

	successPrompt := task.SuccessPrompt()
	require.Contains(t, successPrompt, "my_service.set_task_success")
	require.Contains(t, successPrompt, `project_id="`+project.ID+`"`)
	require.Contains(t, successPrompt, `task_id="`+task.ID+`"`)

	failurePrompt := task.FailurePrompt()
	require.Contains(t, failurePrompt, "my_service.set_task_failure")
	require.Contains(t, failurePrompt, `project_id="`+project.ID+`"`)
	require.Contains(t, failurePrompt, `task_id="`+task.ID+`"`)
}

func TestPlaceholderExpansion(t *testing.T) {
	service := taskcp.New("my_service")
	project := service.AddProject()

	task1 := project.InsertTaskBefore("", "Please complete this task.", "{SUCCESS_PROMPT}", func(project *taskcp.Project, task *taskcp.Task) error { return nil })
	require.Contains(t, task1.Instructions, "my_service.set_task_success")
	require.NotContains(t, task1.Instructions, "{SUCCESS_PROMPT}")

	task2 := project.InsertTaskBefore("", "Try this risky operation.", "{FAILURE_PROMPT}", func(project *taskcp.Project, task *taskcp.Task) error { return nil })
	require.Contains(t, task2.Instructions, "my_service.set_task_failure")
	require.NotContains(t, task2.Instructions, "{FAILURE_PROMPT}")
}

func TestTaskFlow(t *testing.T) {
	service := taskcp.New("test_service")
	project := service.AddProject()

	var completed []string

	task1 := project.InsertTaskBefore("", "First task", "", func(project *taskcp.Project, task *taskcp.Task) error {
		completed = append(completed, task.ID)
		return nil
	})

	task2 := project.InsertTaskBefore("", "Second task", "", func(project *taskcp.Project, task *taskcp.Task) error {
		completed = append(completed, task.ID)
		return nil
	})

	current := project.GetNextTask()
	require.NotNil(t, current)
	require.Equal(t, task1.ID, current.ID)

	next, err := project.SetTaskSuccess(current.ID, "Task 1 done", "")
	require.NoError(t, err)
	require.NotNil(t, next)
	require.Equal(t, task2.ID, next.ID)
	require.Equal(t, taskcp.TaskStateRunning, next.State)

	next2, err := project.SetTaskFailure(next.ID, "Task 2 failed", "Error details")
	require.NoError(t, err)
	require.Nil(t, next2)

	require.Equal(t, []string{task1.ID, task2.ID}, completed)
	require.Equal(t, taskcp.TaskStateSuccess, project.Tasks[task1.ID].State)
	require.Equal(t, taskcp.TaskStateFailure, project.Tasks[task2.ID].State)
}

func TestCallbackError(t *testing.T) {
	service := taskcp.New("test_service")
	project := service.AddProject()

	expectedErr := fmt.Errorf("callback error")

	task := project.InsertTaskBefore("", "Task with error callback", "", func(project *taskcp.Project, task *taskcp.Task) error {
		return expectedErr
	})

	current := project.GetNextTask()
	require.NotNil(t, current)
	require.Equal(t, task.ID, current.ID)

	// Test error propagation on success
	_, err := project.SetTaskSuccess(current.ID, "Result", "")
	require.Error(t, err)
	require.Equal(t, expectedErr, err)

	// Test error propagation on failure
	_, err = project.SetTaskFailure(current.ID, "Task failed", "")
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}
