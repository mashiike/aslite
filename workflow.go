package aslite

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Workflow is configuration of workflow
type Workflow struct {
	Name     string                     `yaml:"name"`
	Defaults *ErrorHandlingStrategy     `yaml:"defaults"`
	Branches map[string]*WorkflowBranch `yaml:"branches"`
}

func (w *Workflow) Validate() error {
	if w.Name == "" {
		return errors.New("workflow name is required")
	}
	if w.Defaults != nil {
		if err := w.Defaults.Validate(); err != nil {
			return err
		}
	}
	if len(w.Branches) == 0 {
		return nil
	}
	for branchName := range w.Branches {
		b := w.Branches[branchName]
		b.Name = branchName
		if err := b.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) NewStateMachine() (*StateMachine, error) {
	sm, err := w.newStateMachine()
	if err != nil {
		return nil, err
	}
	if w.Defaults != nil {
		w.Defaults.ApplyToStateMachine(sm)
	}
	return sm, nil
}

func (w *Workflow) newStateMachine() (*StateMachine, error) {
	sm := &StateMachine{
		States: make(map[string]*State),
	}
	branches := make([]*Branch, 0, len(w.Branches))
	for branchName, branch := range w.Branches {
		b, err := branch.newBranch(branchName)
		if err != nil {
			return nil, err
		}
		if b.Extra == nil {
			b.Extra = make(map[string]any)
		}
		if _, ok := b.Extra["Comment"]; !ok {
			b.Extra["Comment"] = "Branch " + branchName
		}
		branches = append(branches, b)
	}
	if len(branches) == 0 {
		return nil, errors.New("branches is empty")
	}
	if len(branches) == 1 {
		for _, b := range branches {
			sm.States = b.States
			sm.StartAt = b.StartAt
			return sm, nil
		}
	}
	sm.States[w.Name] = &State{
		Type:     "Parallel",
		Branches: branches,
		End:      true,
	}
	sm.StartAt = w.Name
	return sm, nil
}

// WorkflowBranch is configuration of workflow branch
type WorkflowBranch struct {
	Name     string                 `yaml:"-"`
	Defaults *ErrorHandlingStrategy `yaml:"defaults"`
	States   []*WorkflowState       `yaml:"states"`
}

// Validate validates WorkflowBranch
func (wb *WorkflowBranch) Validate() error {
	if wb.Name == "" {
		return errors.New("branch name is required")
	}
	if len(wb.States) == 0 {
		return errors.New("branch states is empty")
	}
	for i, s := range wb.States {
		if s.Name == "" {
			s.Name = fmt.Sprintf("%s State%d", wb.Name, i+1)
		}
		if err := s.Validate(); err != nil {
			return err
		}
	}
	if wb.Defaults != nil {
		if err := wb.Defaults.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (wb *WorkflowBranch) newBranch(branchName string) (*Branch, error) {
	b := &Branch{
		States: make(map[string]*State),
	}
	lastState := &State{
		Type: "Succeed",
	}
	lastStateName := branchName + ":Success"
	b.States[lastStateName] = lastState
	for i := len(wb.States) - 1; i >= 0; i-- {
		s := wb.States[i]
		state, err := s.newState(lastStateName, lastState)
		if err != nil {
			return nil, err
		}
		b.States[s.Name] = state
		lastState = state
		lastStateName = s.Name
	}
	b.StartAt = lastStateName
	if wb.Defaults != nil {
		wb.Defaults.ApplyToBranch(branchName, b)
	}
	return b, nil
}

// WorkflowBranchState is configuration of workflow branch state
type WorkflowState struct {
	Name                  string         `yaml:"name"`
	Uses                  string         `yaml:"uses"`
	With                  map[string]any `yaml:"with"`
	ErrorHandlingStrategy `yaml:",inline"`
	State                 `yaml:"-"`
}

func (ws *WorkflowState) Validate() error {
	if ws.Name == "" {
		return errors.New("state name is required")
	}
	if ws.Uses != "" && ws.State.Type != "" {
		return errors.New("state type and uses are exclusive")
	}
	if err := ws.ErrorHandlingStrategy.Validate(); err != nil {
		return err
	}
	return nil
}

func (ws *WorkflowState) newState(nextStateName string, _ *State) (*State, error) {
	if ws.Uses != "" {
		return nil, errors.New("uses is not supported yet")
	}
	state := ws.State
	state.Next = nextStateName
	return &state, nil
}

func (ws *WorkflowState) UnmarshalYAML(node *yaml.Node) error {
	remakeNode := *node
	remakeNode.Content = make([]*yaml.Node, 0)
	for i := 0; i < len(node.Content); i += 2 {
		switch node.Content[i].Value {
		case "name":
			var str string
			if err := node.Content[i+1].Decode(&str); err != nil {
				return err
			}
			ws.Name = str
		case "uses":
			var str string
			if err := node.Content[i+1].Decode(&str); err != nil {
				return err
			}
			ws.Uses = str
		case "with":
			if err := node.Content[i+1].Decode(&ws.With); err != nil {
				return err
			}
		case "retry":
			if err := node.Content[i+1].Decode(&ws.ErrorHandlingStrategy.Retry); err != nil {
				return err
			}
		case "catch":
			if err := node.Content[i+1].Decode(&ws.ErrorHandlingStrategy.Catch); err != nil {
				return err
			}
		default:
			remakeNode.Content = append(remakeNode.Content, node.Content[i], node.Content[i+1])
		}
	}
	if ws.Uses == "" {
		ws.Type = "Task"
	}
	if err := remakeNode.Decode(&ws.State); err != nil {
		return err
	}
	return nil
}
