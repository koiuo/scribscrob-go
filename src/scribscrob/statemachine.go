package main

import (
	"errors"
	log "github.com/cihub/seelog"
)

// Global types
type (
	State     uint
	InputType uint
	// Function to be called on transition. Doesn't impact StateMachine directly
	SideEffect func()
	/*
		Transition to occur in response to input.

		return next state and list of side effects
	*/
	Transition   func(*Input) (State, []SideEffect)
	InputPayload interface{}
	Input        struct {
		Type    InputType
		Payload InputPayload
	}
	StateMachine struct {
		CurrentState State
		transitions  [][]Transition
	}
)

func NewStateMachine(initialState State, numOfStates uint, numOfInputs uint) *StateMachine {
	preallocatedTransitions := make([][]Transition, numOfStates)
	for i := range preallocatedTransitions {
		preallocatedTransitions[i] = make([]Transition, numOfInputs)
	}

	stateMachine := StateMachine{
		CurrentState: initialState,
		transitions:  preallocatedTransitions,
	}

	return &stateMachine
}

func (stateMachine *StateMachine) AddTransition(state State, inputType InputType, transition Transition) {
	log.Tracef("Adding transition from %d on %d", state, inputType)
	stateMachine.transitions[state][inputType] = transition
}

func (stateMachine *StateMachine) Consume(input *Input) (err error) {
	log.Tracef("Consuming input %v", input)

	state := stateMachine.CurrentState
	inputType := input.Type

	log.Tracef("Getting transition for %d on %d", state, inputType)
	var transition Transition
	transition = stateMachine.transitions[state][inputType]

	if transition != nil {
		nextState, sideEffects := transition(input)
		defer stateMachine.Set(nextState)

		for i, e := range sideEffects {
			log.Tracef("Executing side effect %d...", i)
			e()
			log.Tracef("Side effect %d executed", i)
		}
	} else {
		err = ErrorIllegalInput
	}

	return
}

var (
	ErrorIllegalInput = errors.New("Invalid input")
)

// forcibly set state without invoking any transition functions and causing side-effects
func (stateMachine *StateMachine) Set(state State) {
	stateMachine.CurrentState = state
}
