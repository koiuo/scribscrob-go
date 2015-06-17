package main

import (
	log "github.com/cihub/seelog"
	"os"
	"os/signal"
	"syscall"
)

type Configuration struct {
	// mpd
	MpdAddress  string
	MpdPassword string
	// last.fm
	LastFmUser         string
	LastFmPasswordHash string
}

type Context struct {
	EventBus chan *Input
	Mpd      *Mpd
	Timer    *ThresholdTimer
	LastFm   interface{} // TODO
}

func (ctx *Context) Init(config *Configuration) {
	log.Debug("Initializing context...")
	ctx.EventBus = make(chan *Input, 1)

	// obtain connection to mpd
	mpd := NewMpd(config.MpdAddress, config.MpdPassword, ctx.EventBus)
	mpd.Start()
	ctx.Mpd = mpd

	// play timer
	ctx.Timer = NewThreshodTimer(ctx.EventBus)
	log.Trace("Initialized context")
}

func (ctx *Context) Stop() {
	ctx.Mpd.Stop()
	ctx.Timer.Stop()
	close(ctx.EventBus)
	log.Debug("Stopped context")
}

func main() {
	cfg := &Configuration{
		MpdAddress:"localhost:6600",
		MpdPassword:"",
		LastFmUser:"",
		LastFmPasswordHash:"",
	}
	ctx := &Context{}
	ctx.Init(cfg)

	// state machine
	_ = initstateMachine(ctx)

	// listen for signals until KILL or INT received
	listenSignals(ctx)

	log.Info("Shutting down...")
	ctx.Stop()
}

func listenSignals(ctx *Context) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	for sig := range c {
		log.Infof("Got %s signal", sig.String())
		switch {
		case sig == syscall.SIGHUP:
			log.Info("Reloading Conf...")
			ctx.Stop()
			cfg := &Configuration{
				MpdAddress:"localhost:6600",
				MpdPassword:"",
				LastFmUser:"",
				LastFmPasswordHash:"",
			}
			ctx.Init(cfg)
			break
		case sig == syscall.SIGTERM || sig == syscall.SIGINT:
			return
		}
	}
}

func initstateMachine(ctx *Context) (sm *StateMachine) {
	sm = NewStateMachine(STOPPED, NUM_OF_STATES, NUM_OF_INPUTS)

	// configure transitions for state machine
	sm.AddTransition(PLAYING, PLAY_FROM_START, t(PLAYING, SendNowPlaying(ctx), StartPlayTimer(ctx)))
	sm.AddTransition(PLAYING, PAUSE, t(PAUSED, PausePlayTimer(ctx)))
	sm.AddTransition(PLAYING, STOP, t(STOPPED, StopPlayTimer(ctx)))
	sm.AddTransition(PLAYING, THRESHOLD_REACHED, t(SCROBBLED, Scrobble(ctx)))

	sm.AddTransition(PAUSED, PLAY, t(PLAYING, ContinuePlayTimer(ctx)))
	sm.AddTransition(PAUSED, PLAY_FROM_START, t(PLAYING, SendNowPlaying(ctx), StartPlayTimer(ctx)))
	sm.AddTransition(PAUSED, STOP, t(STOPPED, StopPlayTimer(ctx)))

	sm.AddTransition(STOPPED, PLAY_FROM_START, t(PLAYING, SendNowPlaying(ctx), StartPlayTimer(ctx)))

	sm.AddTransition(SCROBBLED, PLAY_FROM_START, t(PLAYING, SendNowPlaying(ctx), StartPlayTimer(ctx)))
	sm.AddTransition(SCROBBLED, STOP, t(STOPPED))

	go func() {
		for event := range ctx.EventBus {
			sm.Consume(event)
		}
	}()
	return
}
