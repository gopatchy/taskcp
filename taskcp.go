package taskcp

import (
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	Projects   map[string]*Project
	mcpService string
}

type Project struct {
	ID         string
	Tasks      map[string]*Task
	nextTaskID string
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
	ID           string         `json:"id"`
	State        TaskState      `json:"-"`
	Title        string         `json:"title"`
	Instructions string         `json:"instructions"`
	Data         map[string]any `json:"data,omitempty"`
	Result       string         `json:"-"`
	Error        string         `json:"-"`
	Notes        string         `json:"-"`

	NextTaskID string `json:"-"`

	project            *Project
	completionCallback func(project *Project, task *Task) error
}

func New(mcpService string) *Service {
	return &Service{
		Projects:   map[string]*Project{},
		mcpService: mcpService,
	}
}

func (s *Service) AddProject() *Project {
	project := &Project{
		ID:         uuid.New().String(),
		Tasks:      map[string]*Task{},
		nextTaskID: "",
		mcpService: s.mcpService,
	}
	s.Projects[project.ID] = project
	return project
}

func (s *Service) GetProject(id string) (*Project, error) {
	project, ok := s.Projects[id]
	if !ok {
		return nil, fmt.Errorf("project not found")
	}

	return project, nil
}

func (p *Project) InsertTaskBefore(beforeID string, title string, instructions string, completionCallback func(project *Project, task *Task) error) *Task {
	task := p.newTask(title, instructions, completionCallback, beforeID)

	if p.nextTaskID == "" && beforeID == "" {
		p.nextTaskID = task.ID
	} else {
		for t := range p.tasks() {
			if t.NextTaskID == beforeID {
				t.NextTaskID = task.ID
				break
			}
		}
	}

	return task
}

func (p *Project) GetNextTask() *Task {
	if p.nextTaskID == "" {
		return nil
	}

	task := p.Tasks[p.nextTaskID]
	task.State = TaskStateRunning
	return task
}

func (p *Project) SetTaskSuccess(id string, result string, notes string) (*Task, error) {
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

func (p *Project) SetTaskFailure(id string, error string, notes string) (*Task, error) {
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

func (p *Project) newTask(title string, instructions string, completionCallback func(project *Project, task *Task) error, nextTaskID string) *Task {
	task := &Task{
		ID:                 uuid.New().String(),
		State:              TaskStatePending,
		NextTaskID:         nextTaskID,
		Title:              title,
		Instructions:       instructions,
		Data:               map[string]any{},
		completionCallback: completionCallback,
		project:            p,
	}

	task.Instructions = strings.ReplaceAll(task.Instructions, "{SUCCESS_PROMPT}", task.SuccessPrompt())
	task.Instructions = strings.ReplaceAll(task.Instructions, "{FAILURE_PROMPT}", task.FailurePrompt())

	p.Tasks[task.ID] = task
	return task
}

func (p *Project) tasks() iter.Seq[*Task] {
	return func(yield func(*Task) bool) {
		for tid := p.nextTaskID; tid != ""; tid = p.Tasks[tid].NextTaskID {
			t := p.Tasks[tid]
			if !yield(t) {
				return
			}
		}
	}
}

func (t *Task) SuccessPrompt() string {
	return fmt.Sprintf(`To mark this task as successful, use the MCP tool:
%s.set_task_success(project_id="%s", task_id="%s", result="<your result>", notes="<optional notes>")`,
		t.project.mcpService, t.project.ID, t.ID)
}

func (t *Task) FailurePrompt() string {
	return fmt.Sprintf(`To mark this task as failed, use the MCP tool:
%s.set_task_failure(project_id="%s", task_id="%s", error="<error message>", notes="<optional notes>")`,
		t.project.mcpService, t.project.ID, t.ID)
}

func (t *Task) String() string {
	json, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(json)
}
