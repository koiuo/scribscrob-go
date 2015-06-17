package main

import (
	log "github.com/cihub/seelog"
)

// Builds transition function that simply returns next State and slice of SideEffects
func t(to State, with ...func(*Song)) (transition Transition) {
	return func(input *Input) (next State, sideEffects []SideEffect) {
		next = to
		for _, f := range with {
			ff := f
			sideEffects = append(sideEffects, func() {
				ff(input.Payload.(*Song))
			})
		}
		return next, sideEffects
	}
}

// side effects

func StartPlayTimer(ctx *Context) func(song *Song) {
	return func(song *Song) {
		log.Debugf("Starting scrobbling timer for new song %v", song)
		ctx.Timer.Start(song)
	}
}

func PausePlayTimer(ctx *Context) func(song *Song) {
	return func(song *Song) {
		log.Debugf("Pausing scrobbling timer for song %v", song)
		ctx.Timer.Stop()
	}
}

func StopPlayTimer(ctx *Context) func(song *Song) {
	return func(song *Song) {
		log.Debugf("Stopping scrobbling timer for song %v", song)
		ctx.Timer.Stop()
	}
}

func ContinuePlayTimer(ctx *Context) func(song *Song) {
	return func(song *Song) {
		log.Debugf("Resuming scrobbling timer for song %v", song)
		ctx.Timer.Resume()
	}
}

func SendNowPlaying(ctx *Context) func(song *Song) {
	return func(song *Song) {
		log.Debugf("Now playing %v", song)
	}
}

func Scrobble(ctx *Context) func(song *Song) {
	return func(song *Song) {
		log.Debugf("Scrobbling %v", song)
	}
}
