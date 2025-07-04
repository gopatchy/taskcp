package taskcp_test

import (
	"testing"

	"github.com/gopatchy/taskcp"
	"github.com/stretchr/testify/require"
)

func TestTaskCP(t *testing.T) {
	tcp := taskcp.New()

	p := tcp.AddProject()
	require.NotNil(t, p)

	tk := p.InsertTaskBefore(p.NextTaskID, "Hello, world!", func(task *taskcp.Task) {
		t.Logf("Task %s changed: %+v", task.ID, task)
	})
	require.NotNil(t, tk)

	p.SetTaskSuccess(tk.ID, "Hello, world!", "Notes")
	require.Equal(t, taskcp.TaskStateSuccess, tk.State)
	require.Equal(t, "Hello, world!", tk.Result)
	require.Equal(t, "Notes", tk.Notes)
}
