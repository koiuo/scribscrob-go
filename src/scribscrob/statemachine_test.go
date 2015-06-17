package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	A      State = iota
	B            = iota
	C            = iota
	D            = iota
	STATES uint  = iota
)

const (
	X     InputType = iota
	Y               = iota
	Z               = iota
	TYPES uint      = iota
)

func transition(destination State) (simpleTransition Transition) {
	simpleTransition = func(input *Input) (dest State, sideEffects []SideEffect) {
		dest = destination
		sideEffect := func() {
			visibleSideEffect = append(visibleSideEffect, input.Payload)
		}
		sideEffects = []SideEffect{sideEffect}
		return
	}
	return
}

var sm = NewStateMachine(A, STATES, TYPES)
var visibleSideEffect = make([]InputPayload, STATES*TYPES)

func init() {
	sm.AddTransition(A, X, transition(B))
	sm.AddTransition(A, Y, transition(C))
	sm.AddTransition(A, Z, transition(D))

	sm.AddTransition(B, X, transition(A))
	sm.AddTransition(B, Y, transition(C))
	sm.AddTransition(B, Z, transition(D))

	sm.AddTransition(D, X, transition(D))
	sm.AddTransition(D, Y, transition(D))
	sm.AddTransition(D, Z, transition(D))
}

func TestTransition(t *testing.T) {
	sm.Set(A)
	visibleSideEffect = nil

	assert.Nil(t, visibleSideEffect, "Should no be any side effect until transition")

	err := sm.Consume(&Input{
		Type:    X,
		Payload: "AtoB",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, int(sm.CurrentState), B, "Expected transition to B")
		assertSideEffect(t, []InputPayload{"AtoB"})
	}
}

func TestTransitionToSelf(t *testing.T) {
	sm.Set(D)
	visibleSideEffect = nil

	err := sm.Consume(&Input{
		Type:    X,
		Payload: "DonX",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, int(sm.CurrentState), D, "Expected transition to self")
	}

	err = sm.Consume(&Input{
		Type:    Y,
		Payload: "DonY",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, int(sm.CurrentState), D, "Expected transition to self")
	}

	err = sm.Consume(&Input{
		Type:    Y,
		Payload: "DonZ",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, int(sm.CurrentState), D, "Expected transition to self")
	}

	assertSideEffect(t, []InputPayload{"DonX", "DonY", "DonZ"})
}

func TestInvalidTransition(t *testing.T) {
	sm.Set(C)
	visibleSideEffect = nil

	err := sm.Consume(&Input{
		Type:    X,
		Payload: "C invalid",
	})

	if assert.Error(t, err) {
		assertSideEffect(t, []InputPayload{})
	}
}

// private stuff

func assertSideEffect(t *testing.T, expected []InputPayload) {
	actual := visibleSideEffect

	actualLen := len(actual)
	expectedLen := len(expected)
	if expectedLen != actualLen {
		t.Fatalf("Expected slice of length %d but got %d", expectedLen, actualLen)
	}

	for i, v := range expected {
		actualValue := actual[i]
		if v != actualValue {
			t.Fatalf("%d'th element expected %v but got %v", i, v, actualValue)
		}
	}
}
