package policy

import (
	"fmt"
	"strconv"
	"strings"
)

type SpeedPolicy struct {
	speed float32
}

const speedKeyword = "speed"

func init() {
	regFactory(new(speedPolicyFactory))
}

type speedPolicyFactory struct {
}

func (*speedPolicyFactory) Keyword() string {
	return speedKeyword
}

func (*speedPolicyFactory) Build(args []string) (Policy, []string, error) {
	if len(args) == 0 {
		return nil, args, fmt.Errorf("speed should NOT be empty")
	}

	speed, err := parseSpeed(args[0])
	if err != nil {
		return nil, args, err
	}

	return &SpeedPolicy{speed}, args[1:], nil
}

func (s *SpeedPolicy) Keyword() string {
	return speedKeyword
}

func (s *SpeedPolicy) Command() string {
	return speedKeyword + " " + formatSpeed(s.speed)
}

func (s *SpeedPolicy) Comment() string {
	return "匀速 " + formatSpeed(s.speed)
}

func formatSpeed(speed float32) string {
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

func parseSpeed(s string) (float32, error) {
	var times float64 = 1
	s = strings.ToLower(s)
	if strings.HasSuffix(s, "/s") {
		s = s[:len(s)-2]
	}

	if strings.HasSuffix(s, "b") {
		s = s[:len(s)-1]
	}

	if strings.HasSuffix(s, "g") {
		s = s[:len(s)-1]
		times = 1024 * 1024 * 1024
	} else if strings.HasSuffix(s, "m") {
		s = s[:len(s)-1]
		times = 1024 * 1024
	} else if strings.HasSuffix(s, "k") {
		s = s[:len(s)-1]
		times = 1024
	}

	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return -1, err
	} else {
		return float32(f * times), nil
	}
}
