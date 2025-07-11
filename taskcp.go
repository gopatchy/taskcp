package taskcp

import (
	"encoding/json"
	"fmt"
	"iter"
	"strings"
)

type Service struct {
	projects   []*Project
	mcpService string
}

type Project struct {
	ID         int
	Tasks      []*Task
	nextTaskID int
	mcpService string
}

type TaskState string

const (
	TaskStatePending TaskState = "pending"
	TaskStateRunning TaskState = "running"
	TaskStateSuccess TaskState = "success"
	TaskStateFailure TaskState = "failure"
)

type Task struct {
	ID           int            `json:"id"`
	State        TaskState      `json:"-"`
	Title        string         `json:"title"`
	Instructions string         `json:"instructions"`
	Data         map[string]any `json:"data,omitempty"`
	Result       string         `json:"-"`
	Error        string         `json:"-"`
	Notes        string         `json:"-"`

	NextTaskID int `json:"-"`

	project            *Project
	completionCallback func(project *Project, task *Task) error
}

type TaskSummary struct {
	Title string    `json:"title"`
	State TaskState `json:"state"`
	Error string    `json:"error,omitempty"`
	Notes string    `json:"notes,omitempty"`
}

type ProjectSummary struct {
	Tasks []TaskSummary `json:"tasks"`
}

func New(mcpService string) *Service {
	return &Service{
		mcpService: mcpService,
	}
}

func (s *Service) AddProject() *Project {
	project := &Project{
		ID:         len(s.projects),
		Tasks:      []*Task{},
		nextTaskID: -1,
		mcpService: s.mcpService,
	}
	s.projects = append(s.projects, project)
	return project
}

func (s *Service) GetProject(id int) (*Project, error) {
	if id < 0 || id >= len(s.projects) {
		return nil, fmt.Errorf("invalid project id: %d", id)
	}

	return s.projects[id], nil
}

func (p *Project) InsertTaskBefore(beforeID int) *Task {
	newTask := p.newTask(beforeID)

	if p.nextTaskID == -1 && beforeID == -1 {
		p.nextTaskID = newTask.ID
	} else {
		for t := range p.tasks() {
			if t.NextTaskID == beforeID {
				t.NextTaskID = newTask.ID
				break
			}
		}
	}

	return newTask
}

func (p *Project) GetNextTask() *Task {
	if p.nextTaskID == -1 {
		return nil
	}

	task := p.Tasks[p.nextTaskID]
	task.State = TaskStateRunning
	return task
}

func (p *Project) SetTaskSuccess(id int, result string, notes string) (*Task, error) {
	task := p.Tasks[id]
	task.State = TaskStateSuccess
	task.Result = result
	task.Notes = notes

	if task.completionCallback != nil {
		err := task.completionCallback(task.project, task)
		if err != nil {
			return nil, err
		}
	}

	p.nextTaskID = task.NextTaskID

	return p.GetNextTask(), nil
}

func (p *Project) SetTaskFailure(id int, error string, notes string) (*Task, error) {
	task := p.Tasks[id]
	task.State = TaskStateFailure
	task.Error = error
	task.Notes = notes

	if task.completionCallback != nil {
		err := task.completionCallback(task.project, task)
		if err != nil {
			return nil, err
		}
	}

	p.nextTaskID = task.NextTaskID

	return p.GetNextTask(), nil
}

func (p *Project) newTask(nextTaskID int) *Task {
	task := &Task{
		ID:         len(p.Tasks),
		State:      TaskStatePending,
		NextTaskID: nextTaskID,
		Data:       map[string]any{},
		project:    p,
	}

	p.Tasks = append(p.Tasks, task)
	return task
}

func (p *Project) tasks() iter.Seq[*Task] {
	return func(yield func(*Task) bool) {
		for tid := p.nextTaskID; tid != -1; tid = p.Tasks[tid].NextTaskID {
			t := p.Tasks[tid]
			if !yield(t) {
				return
			}
		}
	}
}

func (p *Project) Summary() ProjectSummary {
	var tasks []TaskSummary
	for _, task := range p.Tasks {
		if task.State != TaskStatePending {
			tasks = append(tasks, TaskSummary{
				Title: task.Title,
				State: task.State,
				Error: task.Error,
				Notes: task.Notes,
			})
		}
	}
	return ProjectSummary{Tasks: tasks}
}

func (t *Task) WithTitle(title string) *Task {
	t.Title = title
	return t
}

func (t *Task) WithInstructions(instructions string) *Task {
	t.Instructions = instructions
	t.Instructions = strings.ReplaceAll(t.Instructions, "{SUCCESS_PROMPT}", t.SuccessPrompt())
	t.Instructions = strings.ReplaceAll(t.Instructions, "{FAILURE_PROMPT}", t.FailurePrompt())
	return t
}

func (t *Task) WithData(key string, value any) *Task {
	t.Data[key] = value
	return t
}

func (t *Task) Then(completionCallback func(project *Project, task *Task) error) *Task {
	t.completionCallback = completionCallback
	return t
}

func (t *Task) SuccessPrompt() string {
	return fmt.Sprintf(`To mark this task as successful, use the MCP tool:
%s.set_task_success(project_id=%d, task_id=%d, result="<your result>", notes="<optional notes>")`,
		t.project.mcpService, t.project.ID, t.ID)
}

func (t *Task) FailurePrompt() string {
	return fmt.Sprintf(`To mark this task as failed, use the MCP tool:
%s.set_task_failure(project_id=%d, task_id=%d, error="<error message>", notes="<optional notes>")`,
		t.project.mcpService, t.project.ID, t.ID)
}

func (t *Task) String() string {
	json, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(json)
}

func (ps ProjectSummary) String() string {
	json, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(json)
}
