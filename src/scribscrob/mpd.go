package main

import (
	log "github.com/cihub/seelog"
	"github.com/jteeuwen/go-pkg-mpd"
	"strconv"
	"strings"
	"time"
)

const (
	RECONNECT_TIMEOUT_SECS = 15
)

type Mpd struct {
	client   *mpd.Client
	Address  string
	Password string
	Event    chan *Input
	stopped  bool
	timer    *time.Timer
}

func NewMpd(address, password string, sink chan *Input) (source *Mpd) {
	source = &Mpd{
		Address:  address,
		Password: password,
		Event:    sink,
		stopped:  false,
	}
	return
}

func (s *Mpd) Start() {
	go s.connectAndListen()
}

func (s *Mpd) Stop() {
	log.Debug("Stopping MPD...")
	s.stopped = true
	s.closeResources()
	log.Trace("Stopped MPD")
}

// private methods

func (s *Mpd) connectAndListen() {
	for !s.stopped {
		if s.client == nil {
			s.connect()
			continue
		}

		event, err := s.listenForEvent()
		if err != nil {
			log.Error("Error while idle", err)
			s.closeResources()
			continue
		}
		if event != nil {
			s.Event <- event
		}
	}
}

func (s *Mpd) connect() {
	var client *mpd.Client
	var err error
	for (client == nil) || (err != nil && !s.stopped) {
		log.Debugf("Connecting MPD...")
		client, err = mpd.Dial(s.Address, s.Password)
		if err != nil {
			log.Debugf("Error connecting: %v\n", err)
			s.sleepInterruptibly(RECONNECT_TIMEOUT_SECS * time.Second)
		}
	}

	s.client = client
	log.Info("Established connection to ", s.Address)
}

func (s *Mpd) listenForEvent() (event *Input, err error) {
	changed, err := s.client.IdleSubSystem(mpd.PlayerSystem)
	if err != nil {
		return
	}
	if !changed {
		return
	}

	mpdStatus, err := s.client.Status()
	if err != nil {
		return
	}
	mpdSong, err := s.client.Current()
	if err != nil {
		return
	}

	song := translateMpdSong(mpdSong)
	inputType := translateMpdState(mpdStatus.State, mpdStatus.Elapsed, song)
	event = &Input{
		Type:    inputType,
		Payload: song,
	}
	return
}

func (s *Mpd) closeResources() {
	if (s.client != nil) {
		s.client.Close();
	}
	if (s.timer != nil) {
		s.timer.Stop()
		s.timer = nil
	}
	s.client = nil
}

func (s *Mpd) sleepInterruptibly(duration time.Duration) {
	defer func() {
		s.timer = nil
	}()

	s.timer = time.NewTimer(duration)
	<- s.timer.C
}

// misc

func translateMpdState(state mpd.PlayState, elapsed float32, song *Song) (inputType InputType) {
	switch state {
	case mpd.Playing:
		if (elapsed < 1) || song.IsHttp() {
			inputType = PLAY_FROM_START
		} else {
			inputType = PLAY
		}
		break
	case mpd.Paused:
		inputType = PAUSE
		break
	case mpd.Stopped:
		inputType = STOP
		break
	}
	return
}

func translateMpdSong(mpdSong mpd.Args) (song *Song) {
	source := mpdSong["file"]
	var duration int64
	if !strings.HasPrefix(source, "http") {
		duration, _ = strconv.ParseInt(mpdSong["Time"], 10, 32)
	}
	song = &Song{
		Title:    mpdSong["Title"],
		Artist:   mpdSong["Artist"],
		Album:    mpdSong["Album"],
		duration: time.Duration(duration) * time.Second,
		file:     source,
	}
	return
}
