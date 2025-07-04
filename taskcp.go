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
	ID             string
	State          TaskState
	NextTaskID     string
	ChangeCallback func(task *Task)

	// Written by creator
	Instructions string

	// Written by executor
	Result string
	Error  string
	Notes  string
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

func (p *Project) InsertTaskBefore(id string, instructions string, changeCallback func(task *Task)) *Task {
	task := p.newTask(instructions, changeCallback, id)

	for t := range p.tasks() {
		if t.NextTaskID == id {
			t.NextTaskID = task.ID
			break
		}
	}

	return task
}

func (p *Project) GetNextTask() *Task {
	return p.Tasks[p.NextTaskID]
}

func (p *Project) SetTaskSuccess(id string, result string, notes string) {
	task := p.Tasks[id]
	task.State = TaskStateSuccess
	task.Result = result
	task.Notes = notes
	task.ChangeCallback(task)
	p.NextTaskID = task.NextTaskID
}

func (p *Project) SetTaskFailure(id string, error string, notes string) {
	task := p.Tasks[id]
	task.State = TaskStateFailure
	task.Error = error
	task.Notes = notes
	task.ChangeCallback(task)
}

func (p *Project) newTask(instructions string, changeCallback func(task *Task), nextTaskID string) *Task {
	task := &Task{
		ID:             uuid.New().String(),
		State:          TaskStatePending,
		NextTaskID:     nextTaskID,
		Instructions:   instructions,
		ChangeCallback: changeCallback,
	}
	p.Tasks[task.ID] = task
	return task
}

func (p *Project) tasks() iter.Seq[*Task] {
	return func(yield func(*Task) bool) {
		for tid := p.NextTaskID; tid != ""; tid = p.Tasks[tid].NextTaskID {
			t := p.Tasks[tid]
			if !yield(t) {
				return
			}
		}
	}
}
