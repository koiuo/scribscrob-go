package main

import (
	"time"
)

const (
	// last.fm: minimum song duration is 30s, scrobbling should occur after half is played
	HTTP_SONG_THRESHOLD = time.Duration(15) * time.Second
	MAX_SONG_THRESHOLD  = time.Duration(4) * time.Minute // maximum threshold is 4m
)

// Timer that fires THRESHOLD_REACHED event after song has played for required amount of time
// ThresholdTimer can be stopped and resumed as needed
type ThresholdTimer struct {
	Event       chan *Input
	currentSong *Song
	start       time.Time
	currentTick time.Duration
	targetTick  time.Duration
	timer       *time.Timer
}

func NewThreshodTimer(sink chan *Input) (t *ThresholdTimer) {
	t = &ThresholdTimer{
		Event: sink,
	}
	return
}

func (t *ThresholdTimer) Start(song *Song) {
	t.currentSong = song
	t.currentTick = 0
	t.targetTick = scrobblingThreshold(song)

	if t.timer != nil {
		t.timer.Stop()
	}

	t.timer = time.NewTimer(t.targetTick)
	go t.waitForExpiration()
}

func (t *ThresholdTimer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}

	now := time.Now()
	t.currentTick = now.Sub(t.start)
}

func (t *ThresholdTimer) Resume() {
	duration := t.targetTick - t.currentTick
	t.timer = time.NewTimer(duration)
	go t.waitForExpiration()
}

func (t *ThresholdTimer) waitForExpiration() {
	for _ = range t.timer.C {
		t.Event <- &Input{
			Type:    THRESHOLD_REACHED,
			Payload: t.currentSong,
		}
	}
}

func scrobblingThreshold(song *Song) (threshold time.Duration) {
	if song.IsHttp() {
		threshold = HTTP_SONG_THRESHOLD
	} else {
		threshold = min(MAX_SONG_THRESHOLD, song.duration/2)
	}
	return
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	} else {
		return b
	}
}
