package main

import (
	"strings"
	"time"
)

const (
	PLAY_FROM_START   InputType = iota
	PLAY                        = iota
	PAUSE                       = iota
	STOP                        = iota
	THRESHOLD_REACHED           = iota
	NUM_OF_INPUTS     uint      = iota
)

const (
	PLAYING       State = iota
	PAUSED              = iota
	STOPPED             = iota
	SCROBBLED           = iota
	NUM_OF_STATES uint  = iota
)

type Song struct {
	Title    string
	Artist   string
	Album    string
	file     string
	duration time.Duration
}

func (song *Song) IsHttp() bool {
	return (song != nil) && strings.HasPrefix(song.file, "http")
}
