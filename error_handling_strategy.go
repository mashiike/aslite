package aslite

import (
	"errors"
	"fmt"
)

type ErrorHandlingStrategy struct {
	Retry []*Retrier                      `json:"-" yaml:"retry,omitempty"`
	Catch []*ErrorHandlingStrategyCatcher `json:"-" yaml:"catch,omitempty"`
}

type ErrorHandlingStrategyCatcher struct {
	Name        string   `json:"Name" yaml:"name"`
	ErrorEquals []string `json:"ErrorEquals" yaml:"error_equals,omitempty"`
	ResultPath  string   `json:"ResultPath,omitempty" yaml:"result_path,omitempty"`
	Comment     string   `json:"Comment,omitempty" yaml:"comment,omitempty"`
	SNSTopicARN string   `json:"SNSTopicArn" yaml:"sns_topic_arn,omitempty"`
}

func (ehs *ErrorHandlingStrategy) Validate() error {
	for _, r := range ehs.Retry {
		if len(r.ErrorEquals) == 0 {
			r.ErrorEquals = []string{"States.ALL"}
		}
	}
	for i, c := range ehs.Catch {
		if c.Name == "" {
			c.Name = fmt.Sprintf("DefaultCatcher%d", i+1)
		}
		if len(c.ErrorEquals) == 0 {
			c.ErrorEquals = []string{"States.ALL"}
		}
	}
	return nil
}

func (ehsc *ErrorHandlingStrategyCatcher) NewStates() map[string]*State {
	if ehsc.SNSTopicARN != "" {
		return map[string]*State{
			ehsc.Name: {
				Type:    "Task",
				Comment: ehsc.Comment,
				Next:    ehsc.Name + ":Faild",
				Extra: map[string]any{
					"Resource": "arn:aws:states:::sns:publish",
					"Parameters": map[string]any{
						"TopicArn":  ehsc.SNSTopicARN,
						"Message.$": "$",
					},
				},
			},
			ehsc.Name + ":Faild": {
				Type:    "Fail",
				Comment: ehsc.Comment,
			},
		}
	}
	return map[string]*State{
		ehsc.Name: {
			Type:    "Fail",
			Comment: ehsc.Comment,
		},
	}
}

func (ehs *ErrorHandlingStrategy) ApplyToStateMachine(sm *StateMachine) error {
	needCacherNames := make(map[string]struct{})
	for stateName, state := range sm.States {
		names, err := ehs.ApplyToState(stateName, state)
		if err != nil {
			return err
		}
		for _, name := range names {
			needCacherNames[name] = struct{}{}
		}
	}
	for _, cacher := range ehs.Catch {
		if _, ok := needCacherNames[cacher.Name]; !ok {
			continue
		}
		for name, state := range cacher.NewStates() {
			if _, ok := sm.States[name]; ok {
				return errors.New("state name conflict: " + name)
			}
			sm.States[name] = state
		}
	}
	return nil
}

func (ehs *ErrorHandlingStrategy) ApplyToState(stateName string, s *State) ([]string, error) {
	switch s.Type {
	case "Fail", "Succeed", "Choice", "Wait", "Pass":
		return nil, nil
	case "Task":
		if len(s.Retry) == 0 {
			s.Retry = ehs.Retry
		}
		if len(s.Catch) == 0 {
			var needCacherNames []string
			s.Catch = make([]*Catcher, len(ehs.Catch))
			for i, c := range ehs.Catch {
				s.Catch[i] = &Catcher{
					ErrorEquals: c.ErrorEquals,
					ResultPath:  c.ResultPath,
					Next:        c.Name,
				}
				extra := map[string]any{}
				if c.Comment != "" {
					extra["Comment"] = c.Comment
				}
				s.Catch[i].Extra = extra
				needCacherNames = append(needCacherNames, c.Name)
			}
			return needCacherNames, nil
		}
		return nil, nil
	}
	if s.Branches != nil {
		for i, branch := range s.Branches {
			branchName := fmt.Sprintf("%s.Branches[%d]", stateName, i)
			if err := ehs.ApplyToBranch(branchName, branch); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	if s.Iterator != nil {
		branchName := stateName + ".Iterator"
		if err := ehs.ApplyToBranch(branchName, s.Iterator); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (ehs *ErrorHandlingStrategy) ApplyToBranch(branchName string, b *Branch) error {
	cachers := make([]*ErrorHandlingStrategyCatcher, len(ehs.Catch))
	for i, c := range ehs.Catch {
		cachers[i] = &ErrorHandlingStrategyCatcher{
			Name:        fmt.Sprintf("%s.%s", branchName, c.Name),
			ErrorEquals: c.ErrorEquals,
			ResultPath:  c.ResultPath,
			Comment:     c.Comment,
			SNSTopicARN: c.SNSTopicARN,
		}
	}
	needCacherNames := make(map[string]struct{})
	for stateName, state := range b.States {
		names, err := ehs.ApplyToState(stateName, state)
		if err != nil {
			return err
		}
		for _, name := range names {
			needCacherNames[name] = struct{}{}
		}
	}
	for _, cacher := range cachers {
		if _, ok := needCacherNames[cacher.Name]; !ok {
			continue
		}
		for name, state := range cacher.NewStates() {
			if _, ok := b.States[name]; ok {
				return errors.New("state name conflict: " + name)
			}
			b.States[name] = state
		}
	}
	return nil
}
