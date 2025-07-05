package taskcp

import (
	"fmt"
	"iter"

	"github.com/google/uuid"
)

type Service struct {
	Projects map[string]*Project
}

type Project struct {
	ID         string
	Tasks      map[string]*Task
	NextTaskID string
}

type TaskState string

const (
	TaskStatePending TaskState = "pending"
	TaskStateRunning TaskState = "running"
	TaskStateSuccess TaskState = "success"
	TaskStateFailure TaskState = "failure"
)

type Task struct {
	ID           string    `json:"id"`
	State        TaskState `json:"-"`
	Instructions string    `json:"instructions"`
	Result       string    `json:"-"`
	Error        string    `json:"-"`
	Notes        string    `json:"-"`

	nextTaskID         string
	completionCallback func(task *Task)
}

func New() *Service {
	return &Service{
		Projects: map[string]*Project{},
	}
}

func (s *Service) AddProject() *Project {
	project := &Project{
		ID:         uuid.New().String(),
		Tasks:      map[string]*Task{},
		NextTaskID: "",
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

func (p *Project) InsertTaskBefore(beforeID string, instructions string, completionCallback func(task *Task)) *Task {
	task := p.newTask(instructions, completionCallback, beforeID)

	for t := range p.tasks() {
		if t.nextTaskID == beforeID {
			t.nextTaskID = task.ID
			break
		}
	}

	return task
}

func (p *Project) GetNextTask() *Task {
	if p.NextTaskID == "" {
		return nil
	}

	task := p.Tasks[p.NextTaskID]
	task.State = TaskStateRunning
	return task
}

func (p *Project) SetTaskSuccess(id string, result string, notes string) *Task {
	task := p.Tasks[id]
	task.State = TaskStateSuccess
	task.Result = result
	task.Notes = notes
	task.completionCallback(task)
	p.NextTaskID = task.nextTaskID

	return p.GetNextTask()
}

func (p *Project) SetTaskFailure(id string, error string, notes string) *Task {
	task := p.Tasks[id]
	task.State = TaskStateFailure
	task.Error = error
	task.Notes = notes
	task.completionCallback(task)
	p.NextTaskID = task.nextTaskID

	return p.GetNextTask()
}

func (p *Project) newTask(instructions string, completionCallback func(task *Task), nextTaskID string) *Task {
	task := &Task{
		ID:                 uuid.New().String(),
		State:              TaskStatePending,
		nextTaskID:         nextTaskID,
		Instructions:       instructions,
		completionCallback: completionCallback,
	}
	p.Tasks[task.ID] = task
	return task
}

func (p *Project) tasks() iter.Seq[*Task] {
	return func(yield func(*Task) bool) {
		for tid := p.NextTaskID; tid != ""; tid = p.Tasks[tid].nextTaskID {
			t := p.Tasks[tid]
			if !yield(t) {
				return
			}
		}
	}
}
