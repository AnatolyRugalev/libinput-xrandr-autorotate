package autorotate

import (
	"context"
	"errors"
	"github.com/AnatolyRugalev/libinput-xrandr-autorotate/pkg/accelerometer"
	"github.com/AnatolyRugalev/libinput-xrandr-autorotate/pkg/exec"
	"strconv"
	"strings"
	"time"
)

type Orientation string
type Axis int
type XInputCoordinates [9]int

const (
	OrientationNormal   = Orientation("normal")
	OrientationInverted = Orientation("inverted")
	OrientationLeft     = Orientation("left")
	OrientationRight    = Orientation("right")

	AxisX = Axis(iota)
	AxisY
)

var orientationMap = map[Orientation]XInputCoordinates{
	OrientationNormal:   {1, 0, 0, 0, 1, 0, 0, 0, 1},
	OrientationInverted: {-1, 0, 1, 0, -1, 1, 0, 0, 1},
	OrientationLeft:     {0, -1, 1, 1, 0, 0, 0, 0, 1},
	OrientationRight:    {0, 1, 0, -1, 0, 1, 0, 0, 1},
}

var AutodetectKeywords = []string{
	"Wacom HID",
}

type edge struct {
	axis Axis
	min  float64
	max  float64
}

func (a Autorotate) SetOrientation(orientation Orientation) error {
	if err := xrandrCommand(orientation, a.display); err != nil {
		return err
	}

	for _, touchscreen := range a.touchscreens {
		if err := xinputCommand(orientation, touchscreen); err != nil {
			return err
		}
	}
	return nil
}

func xrandrCommand(orientation Orientation, display string) error {
	_, err := exec.ExecuteCommand("xrandr", "-d", display, "-o", string(orientation))
	return err
}

func xinputCommand(orientation Orientation, touchscreen string) error {
	matrix := xinputString(orientationMap[orientation])
	args := []string{"set-prop", touchscreen, "Coordinate Transformation Matrix"}
	args = append(args, matrix...)
	_, err := exec.ExecuteCommand("xinput", args...)
	return err
}

func xinputString(coords XInputCoordinates) []string {
	var strs []string
	for _, val := range coords {
		strs = append(strs, strconv.Itoa(val))
	}
	return strs
}

func DetectTouchScreens() ([]string, error) {
	namesStr, err := exec.ExecuteCommand("xinput", "list", "--name-only")
	if err != nil {
		return nil, err
	}
	var screens []string
	names := strings.Split(namesStr, "\n")
	for _, name := range names {
		for _, kw := range AutodetectKeywords {
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

func (a Autorotate) GetOrientationEdges() map[Orientation]edge {
	return map[Orientation]edge{
		OrientationNormal: {
			axis: AxisY,
			min:  -100.0,
			max:  -a.threshold,
		},
		OrientationInverted: {
			axis: AxisY,
			min:  a.threshold,
			max:  100.0,
		},
		OrientationLeft: {
			axis: AxisX,
			min:  a.threshold,
			max:  100.0,
		},
		OrientationRight: {
			axis: AxisX,
			min:  -100.0,
			max:  -a.threshold,
		},
	}
}

func (a *Autorotate) Watch(ctx context.Context) error {
	reader := accelerometer.NewReader(a.accelerometer)
	err := reader.Init()
	if err != nil {
		return err
	}
	vals := make(chan accelerometer.Value)
	go reader.Read(ctx, a.refreshRate, vals)
	a.state = &state{
		autorotate:  a,
		orientation: OrientationNormal,
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case val := <-vals:
			a.state.update(val)
		}
	}
}

type Autorotate struct {
	display       string
	touchscreens  []string
	accelerometer string
	threshold     float64
	refreshRate   time.Duration
	maxTicks      int
	state         *state
}

func NewAutorotate(display string, touchscreens []string, accelerometerName string, threshold float64, refreshRate time.Duration, maxTicks int) *Autorotate {
	return &Autorotate{
		display:       display,
		touchscreens:  touchscreens,
		accelerometer: accelerometerName,
		threshold:     threshold,
		refreshRate:   refreshRate,
		maxTicks:      maxTicks,
	}
}
