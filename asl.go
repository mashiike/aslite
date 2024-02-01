package aslite

import (
	"reflect"

	"github.com/serenize/snaker"
	"gopkg.in/yaml.v3"
)

type StateMachine struct {
	States  map[string]*State      `json:"States"`
	StartAt string                 `json:"StartAt"`
	Comment string                 `json:"Comment,omitempty"`
	Extra   map[string]interface{} `json:"-"`
}

func (asm *StateMachine) UnmarshalJSON(data []byte) error {
	type Alias StateMachine
	aux := &struct {
		*Alias `json:",inline"`
	}{
		Alias: (*Alias)(asm),
	}
	extra, err := UnmarshalJSONWithExtra(data, aux)
	if err != nil {
		return err
	}
	asm.Extra = extra
	return nil
}

func (asm *StateMachine) MarshalJSON() ([]byte, error) {
	type Alias StateMachine
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(asm),
	}
	return MarshalJSONWithExtra(aux, asm.Extra)
}

func (asm *StateMachine) OnlyParalell() bool {
	for _, state := range asm.States {
		if state.Type != "Parallel" {
			return false
		}
	}
	return true
}

func (asm *StateMachine) IsLinearWorkflow() bool {
	return isLinearWorkflow(asm.States, true)
}

func onlyParalell(states map[string]*State) bool {
	for _, state := range states {
		if state.Type != "Parallel" {
			return false
		}
	}
	return true
}

func isLinearWorkflow(states map[string]*State, checkParalell bool) bool {
	if checkParalell && onlyParalell(states) {
		for _, state := range states {
			for _, branch := range state.Branches {
				if !isLinearWorkflow(branch.States, false) {
					return false
				}
			}
		}
		return true
	}
	// no have branch and iterator
	dependency := make(map[string]string, len(states))
	choices := make(map[string]string, len(states))
	for stateName, state := range states {
		if len(state.Branches) > 0 || state.Iterator != nil {
			return false
		}
		if len(state.Choices) > 0 {
			if len(state.Choices) > 1 {
				return false
			}
			choices[stateName] = state.Choices[0].Next
			dependency[stateName] = state.Default
		} else {
			if state.Next != "" {
				dependency[stateName] = state.Next
			} else {
				dependency[stateName] = ""
			}
		}
	}
	for c, d := range choices {
		defaultRoute, ok := dependency[c]
		if !ok {
			continue
		}
		ifRoute, ok := dependency[d]
		if !ok {
			continue
		}
		if defaultRoute == ifRoute {
			continue
		}
		return false
	}
	return true
}

type State struct {
	// common fields
	Type    string     `json:"Type" yaml:"type"`
	Comment string     `json:"Comment,omitempty" yaml:"comment,omitempty"`
	Next    string     `json:"Next,omitempty" yaml:"-"`
	End     bool       `json:"End,omitempty" yaml:"-"`
	Retry   []*Retrier `json:"Retry,omitempty" yaml:"-"`
	Catch   []*Catcher `json:"Catch,omitempty" yaml:"-"`

	// Choice State specific fields
	Choices []*ChoiceRule `json:"Choices,omitempty" yaml:"-"`
	Default string        `json:"Default,omitempty" yaml:"-"`

	// Parallel State specific fields
	Branches []*Branch `json:"Branches,omitempty" yaml:"-"`

	// Map State specific fields
	Iterator *Branch `json:"Iterator,omitempty" yaml:"-"`

	// Fail State specific fields
	Cause     string `json:"Cause,omitempty" yaml:"cause,omitempty"`
	Error     string `json:"Error,omitempty" yaml:"error,omitempty"`
	CausePath string `json:"CausePath,omitempty" yaml:"cause_path,omitempty"`
	ErrorPath string `json:"ErrorPath,omitempty" yaml:"error_path,omitempty"`

	// other fields
	Extra map[string]interface{} `json:"-" yaml:"-"`
}

func (as *State) UnmarshalJSON(data []byte) error {
	type Alias State
	aux := &struct {
		*Alias `json:",inline"`
	}{
		Alias: (*Alias)(as),
	}
	extra, err := UnmarshalJSONWithExtra(data, aux)
	if err != nil {
		return err
	}
	as.Extra = extra
	return nil
}

func (as *State) UnmarshalYAML(node *yaml.Node) error {
	type Alias State
	aux := &struct {
		*Alias `yaml:",inline"`
	}{
		Alias: (*Alias)(as),
	}
	if err := node.Decode(aux); err != nil {
		return err
	}
	rv := reflect.ValueOf(as)
	t := rv.Type()
	fields := make(map[string]struct{})
	extractFields(t, fields, "", "yaml")

	var raw map[string]interface{}
	if err := node.Decode(&raw); err != nil {
		return err
	}

	extra := make(map[string]interface{})
	for k, v := range raw {
		if _, ok := fields[k]; !ok {
			newKey := snaker.SnakeToCamel(k)
			extra[newKey] = v
		}
	}
	as.Extra = extra
	return nil
}

func (as *State) MarshalJSON() ([]byte, error) {
	type Alias State
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(as),
	}
	return MarshalJSONWithExtra(aux, as.Extra)
}

type ChoiceRule struct {
	Variable string `json:"Variable,omitempty"`
	Next     string `json:"Next"`

	And []*ChoiceRule `json:"And,omitempty"`
	Or  []*ChoiceRule `json:"Or,omitempty"`
	Not *ChoiceRule   `json:"Not,omitempty"`

	Comment string                 `json:"Comment,omitempty"`
	Extra   map[string]interface{} `json:"-"`
}

func (acr *ChoiceRule) UnmarshalJSON(data []byte) error {
	type Alias ChoiceRule
	aux := &struct {
		*Alias `json:",inline"`
	}{
		Alias: (*Alias)(acr),
	}
	extra, err := UnmarshalJSONWithExtra(data, aux)
	if err != nil {
		return err
	}
	acr.Extra = extra
	return nil
}

func (acr ChoiceRule) MarshalJSON() ([]byte, error) {
	type Alias ChoiceRule
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(&acr),
	}
	return MarshalJSONWithExtra(aux, acr.Extra)
}

type Branch struct {
	States  map[string]*State      `json:"States"`
	StartAt string                 `json:"StartAt"`
	Extra   map[string]interface{} `json:"-"`
}

func (b *Branch) UnmarshalJSON(data []byte) error {
	type Alias Branch
	aux := &struct {
		*Alias `json:",inline"`
	}{
		Alias: (*Alias)(b),
	}
	extra, err := UnmarshalJSONWithExtra(data, aux)
	if err != nil {
		return err
	}
	b.Extra = extra
	return nil
}

func (b *Branch) MarshalJSON() ([]byte, error) {
	type Alias Branch
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	return MarshalJSONWithExtra(aux, b.Extra)
}

type Retrier struct {
	ErrorEquals     []string               `json:"ErrorEquals" yaml:"error_equals,omitempty"`
	IntervalSeconds int                    `json:"IntervalSeconds,omitempty" yaml:"interval_seconds,omitempty"`
	MaxAttempts     int                    `json:"MaxAttempts,omitempty" yaml:"max_attempts,omitempty"`
	BackoffRate     float64                `json:"BackoffRate,omitempty" yaml:"backoff_rate,omitempty"`
	MaxDelaySeconds int                    `json:"MaxDelaySeconds,omitempty" yaml:"max_delay_seconds,omitempty"`
	JitterStrategy  string                 `json:"JitterStrategy,omitempty" yaml:"jitter_strategy,omitempty"`
	Extra           map[string]interface{} `json:"-"`
}

func (ar *Retrier) UnmarshalJSON(data []byte) error {
	type Alias Retrier
	aux := &struct {
		*Alias `json:",inline"`
	}{
		Alias: (*Alias)(ar),
	}
	extra, err := UnmarshalJSONWithExtra(data, aux)
	if err != nil {
		return err
	}
	ar.Extra = extra
	return nil
}

func (ar *Retrier) MarshalJSON() ([]byte, error) {
	type Alias Retrier
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(ar),
	}
	return MarshalJSONWithExtra(aux, ar.Extra)
}

type Catcher struct {
	ErrorEquals []string               `json:"ErrorEquals"`
	ResultPath  string                 `json:"ResultPath,omitempty"`
	Next        string                 `json:"Next"`
	Extra       map[string]interface{} `json:"-"`
}

func (ac *Catcher) UnmarshalJSON(data []byte) error {
	type Alias Catcher
	aux := &struct {
		*Alias `json:",inline"`
	}{
		Alias: (*Alias)(ac),
	}
	extra, err := UnmarshalJSONWithExtra(data, aux)
	if err != nil {
		return err
	}
	ac.Extra = extra
	return nil
}

func (ac *Catcher) MarshalJSON() ([]byte, error) {
	type Alias Catcher
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(ac),
	}
	return MarshalJSONWithExtra(aux, ac.Extra)
}
