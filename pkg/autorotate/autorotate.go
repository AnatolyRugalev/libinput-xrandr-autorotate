package autorotate

import (
	"errors"
	"fmt"
	"github.com/AnatolyRugalev/libinput-xrandr-autorotate/pkg/exec"
	"strconv"
	"strings"
	"time"
)

type Orientation string
type LibInputCoordinates [9]int

const (
	OrientationNormal   = Orientation("normal")
	OrientationInverted = Orientation("inverted")
	OrientationLeft     = Orientation("left")
	OrientationRight    = Orientation("right")
)

var orientationMap = map[Orientation]LibInputCoordinates{
	OrientationNormal:   {1, 0, 0, 0, 1, 0, 0, 0, 1},
	OrientationInverted: {-1, 0, 1, 0, -1, 1, 0, 0, 1},
	OrientationLeft:     {0, -1, 1, 1, 0, 0, 0, 0, 1},
	OrientationRight:    {0, 1, 0, -1, 0, 1, 0, 0, 1},
}

var autodetectKeywords = []string{
	"Wacom HID",
}

func SetOrientation(devices []string, display string, orientation Orientation) error {
	if err := xrandrCommand(orientation, display); err != nil {
		return err
	}

	for _, device := range devices {
		if err := xinputCommand(orientation, device); err != nil {
			return err
		}
	}
	return nil
}

func xrandrCommand(orientation Orientation, display string) error {
	_, err := exec.ExecuteCommand("xrandr", "-d", display, "-o", string(orientation))
	return err
}

func xinputCommand(orientation Orientation, device string) error {
	matrix := CoordinatesToString(orientationMap[orientation])
	args := []string{"set-prop", device, "Coordinate Transformation Matrix"}
	args = append(args, matrix...)
	_, err := exec.ExecuteCommand("xinput", args...)
	return err
}

func CoordinatesToString(coords LibInputCoordinates) []string {
	var strs []string
	for _, val := range coords {
		strs = append(strs, strconv.Itoa(val))
	}
	return strs
}

func GetTouchScreens() ([]string, error) {
	namesStr, err := exec.ExecuteCommand("xinput", "list", "--name-only")
	if err != nil {
		return nil, err
	}
	var screens []string
	names := strings.Split(namesStr, "\n")
	for _, name := range names {
		for _, kw := range autodetectKeywords {
			if strings.Contains(name, kw) {
				screens = append(screens, name)
				break
			}
		}
	}
	if len(screens) == 0 {
		return nil, errors.New("no touchscreens found")
	}
	return screens, err
}

type value struct {
	x float64
	y float64
	z float64
}

type state struct {
	orientation    Orientation
	newOrientation *Orientation
	display        string
	touchscreens   []string
	ticks          int
	maxTicks       int
	threshold      float64
}

func Watch(exit <-chan struct{}, display string, touchscreens []string, accelerometer string, threshold float64, refreshRate time.Duration, ticks int) error {
	vals, err := ReadValues(accelerometer, exit, refreshRate)
	if err != nil {
		return err
	}
	state := &state{
		orientation:  OrientationNormal,
		display:      display,
		touchscreens: touchscreens,
		maxTicks:     ticks,
		threshold:    threshold,
	}
	for {
		select {
		case <-exit:
			return nil
		case val := <-vals:
			state.update(val)
		}
	}
}

type edge struct {
	y   bool
	min float64
	max float64
}

func GetOrientationEdges(threshold float64) map[Orientation]edge {
	return map[Orientation]edge{
		OrientationNormal: {
			y:   true,
			min: -100.0,
			max: -threshold,
		},
		OrientationInverted: {
			y:   true,
			min: threshold,
			max: 100.0,
		},
		OrientationLeft: {
			y:   false,
			min: threshold,
			max: 100.0,
		},
		OrientationRight: {
			y:   false,
			min: -100.0,
			max: -threshold,
		},
	}
}

func (s state) detectOrientation(val value) Orientation {
	for o, threshold := range GetOrientationEdges(s.threshold) {
		if o == s.orientation {
			continue
		}
		if threshold.y {
			if val.y >= threshold.min &&
				val.y < threshold.max {
				return o
			}
		} else {
			if val.x >= threshold.min &&
				val.x < threshold.max {
				return o
			}
		}
	}
	return s.orientation
}

func (s *state) update(val value) {
	o := s.detectOrientation(val)
	//fmt.Printf("x = %.4f y = %.4f z = %.4f o = %s\n", val.x, val.y, val.z, o)
	if o != s.orientation {
		if s.newOrientation == nil || *s.newOrientation != o {
			s.newOrientation = &o
			s.ticks = 0
		} else {
			s.ticks++
			if s.ticks > s.maxTicks {
				s.orientation = o
				s.newOrientation = nil
				s.ticks = 0
				err := SetOrientation(s.touchscreens, s.display, o)
				if err != nil {
					fmt.Printf("Error changing orientation: %s\n", err.Error())
				}
			}
		}
	}
}
