package taskcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Service struct {
	projects   []*Project
	mcpService string
}

type Project struct {
	ID           int
	PendingTasks []*Task
	RunningTasks []*Task
	SuccessTasks []*Task
	FailureTasks []*Task
	mcpService   string
	nextTaskID   int
}

type Task struct {
	ID           int            `json:"id"`
	Title        string         `json:"title"`
	Instructions string         `json:"instructions"`
	Data         map[string]any `json:"data,omitempty"`
	Result       string         `json:"-"`
	Error        string         `json:"-"`
	Notes        string         `json:"-"`

	completionCallback func(task *Task) error
	project            *Project
}

type TaskSummary struct {
	Title string `json:"title"`
	Error string `json:"error,omitempty"`
	Notes string `json:"notes,omitempty"`
}

type ProjectSummary struct {
	PendingTasks []*TaskSummary `json:"pending_tasks"`
	RunningTasks []*TaskSummary `json:"running_tasks"`
	SuccessTasks []*TaskSummary `json:"success_tasks"`
	FailureTasks []*TaskSummary `json:"failure_tasks"`
}

func New(mcpService string) *Service {
	return &Service{
		mcpService: mcpService,
	}
}

func (s *Service) AddProject() *Project {
	project := &Project{
		ID:           len(s.projects),
		PendingTasks: []*Task{},
		RunningTasks: []*Task{},
		SuccessTasks: []*Task{},
		FailureTasks: []*Task{},
		mcpService:   s.mcpService,
		nextTaskID:   0,
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

func (p *Project) AddNextTask() *Task {
	t := p.newTask()
	p.PendingTasks = append([]*Task{t}, p.PendingTasks...)
	return t
}

func (p *Project) AddLastTask() *Task {
	t := p.newTask()
	p.PendingTasks = append(p.PendingTasks, t)
	return t
}

func (p *Project) newTask() *Task {
	task := &Task{
		ID:      p.nextTaskID,
		Data:    map[string]any{},
		project: p,
	}

	p.nextTaskID++

	return task
}

func (p *Project) PopNextTask() (*Task, error) {
	if len(p.PendingTasks) == 0 {
		return nil, nil
	}

	task := p.PendingTasks[0]
	p.PendingTasks = p.PendingTasks[1:]
	return task, nil
}

func (p *Project) GetRunningTask(id int) (*Task, error) {
	for _, task := range p.RunningTasks {
		if task.ID == id {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %d", id)
}

func (p *Project) Summary() *ProjectSummary {
	s := &ProjectSummary{
		PendingTasks: []*TaskSummary{},
		RunningTasks: []*TaskSummary{},
		SuccessTasks: []*TaskSummary{},
		FailureTasks: []*TaskSummary{},
	}

	for _, task := range p.PendingTasks {
		s.PendingTasks = append(s.PendingTasks, task.AsSummary())
	}

	for _, task := range p.RunningTasks {
		s.RunningTasks = append(s.RunningTasks, task.AsSummary())
	}

	for _, task := range p.SuccessTasks {
		s.SuccessTasks = append(s.SuccessTasks, task.AsSummary())
	}

	for _, task := range p.FailureTasks {
		s.FailureTasks = append(s.FailureTasks, task.AsSummary())
	}

	return s
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

func (t *Task) Then(completionCallback func(task *Task) error) {
	t.completionCallback = completionCallback
}

func (t *Task) SetSuccess(result string, notes string) (*Task, error) {
	t.Result = result
	t.Notes = notes

	t.project.SuccessTasks = append(t.project.SuccessTasks, t)

	if t.completionCallback != nil {
		err := t.completionCallback(t)
		if err != nil {
			return nil, err
		}
	}

	return t.project.PopNextTask()
}

func (t *Task) SetFailure(error string, notes string) (*Task, error) {
	t.Error = error
	t.Notes = notes

	t.project.FailureTasks = append(t.project.FailureTasks, t)

	if t.completionCallback != nil {
		err := t.completionCallback(t)
		if err != nil {
			return nil, err
		}
	}

	return t.project.PopNextTask()
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

func (t *Task) AsSummary() *TaskSummary {
	return &TaskSummary{
		Title: t.Title,
		Error: t.Error,
		Notes: t.Notes,
	}
}

func (ps ProjectSummary) String() string {
	json, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(json)
}
