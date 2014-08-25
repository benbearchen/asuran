package profile

import (
	"strconv"
)

type SpeedActType int

const (
	SpeedActNone = iota
	SpeedActConstant
)

type SpeedAction struct {
	Act   SpeedActType
	Speed float32
}

func (s *SpeedAction) SpeedCommand() string {
	return s.speedString(s.Speed)
}

func (s *SpeedAction) speedString(speed float32) string {
	if speed >= 1024*1024*1024 {
		return strconv.FormatFloat(float64(speed)/1024/1024/1024, 'f', -1, 32) + "GB/s"
	} else if speed >= 1024*1024 {
		return strconv.FormatFloat(float64(speed)/1024/1024, 'f', -1, 32) + "MB/s"
	} else if speed >= 1024 {
		return strconv.FormatFloat(float64(speed)/1024, 'f', -1, 32) + "KB/s"
	} else {
		return strconv.FormatFloat(float64(speed), 'f', -1, 32) + "B/s"
	}
}

func (s *SpeedAction) String() string {
	switch s.Act {
	case SpeedActConstant:
		return "匀速 " + s.SpeedCommand()
	default:
		return ""
	}
}

func MakeEmptySpeed() SpeedAction {
	return SpeedAction{SpeedActNone, 0}
}

func MakeSpeed(act SpeedActType, speed float32) SpeedAction {
	return SpeedAction{act, speed}
}
