package taskcp_test

import (
	"testing"

	"github.com/gopatchy/taskcp"
	"github.com/stretchr/testify/require"
)


func TestTaskPrompts(t *testing.T) {
	service := taskcp.New("my_service")
	project := service.AddProject()
	
	task := project.InsertTaskBefore("", "Write unit tests", func(task *taskcp.Task) {})
	
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
	
	task1 := project.InsertTaskBefore("", "Please complete this task. {SUCCESS_PROMPT}", func(task *taskcp.Task) {})
	require.Contains(t, task1.Instructions, "my_service.set_task_success")
	require.NotContains(t, task1.Instructions, "{SUCCESS_PROMPT}")
	
	task2 := project.InsertTaskBefore("", "Try this risky operation. {FAILURE_PROMPT}", func(task *taskcp.Task) {})
	require.Contains(t, task2.Instructions, "my_service.set_task_failure")
	require.NotContains(t, task2.Instructions, "{FAILURE_PROMPT}")
}

func TestTaskFlow(t *testing.T) {
	service := taskcp.New("test_service")
	project := service.AddProject()
	
	var completed []string
	
	task1 := project.InsertTaskBefore("", "First task", func(task *taskcp.Task) {
		completed = append(completed, task.ID)
	})
	
	task2 := project.InsertTaskBefore("", "Second task", func(task *taskcp.Task) {
		completed = append(completed, task.ID)
	})
	
	current := project.GetNextTask()
	require.NotNil(t, current)
	require.Equal(t, task1.ID, current.ID)
	
	next := project.SetTaskSuccess(current.ID, "Task 1 done", "")
	require.NotNil(t, next)
	require.Equal(t, task2.ID, next.ID)
	require.Equal(t, taskcp.TaskStateRunning, next.State)
	
	next2 := project.SetTaskFailure(next.ID, "Task 2 failed", "Error details")
	require.Nil(t, next2)
	
	require.Equal(t, []string{task1.ID, task2.ID}, completed)
	require.Equal(t, taskcp.TaskStateSuccess, project.Tasks[task1.ID].State)
	require.Equal(t, taskcp.TaskStateFailure, project.Tasks[task2.ID].State)
}