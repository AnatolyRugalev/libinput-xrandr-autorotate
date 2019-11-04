package autorotate

import (
	"fmt"
	"github.com/AnatolyRugalev/libinput-xrandr-autorotate/pkg/accelerometer"
)

type state struct {
	autorotate     *Autorotate
	orientation    Orientation
	newOrientation *Orientation
	ticks          int
}

func (s state) detectOrientation(val accelerometer.Value) Orientation {
	for o, edge := range s.autorotate.GetOrientationEdges() {
		if o == s.orientation {
			continue
		}
		if edge.axis == AxisY {
			if val.Y >= edge.min &&
				val.Y < edge.max {
				return o
			}
		} else {
			if val.X >= edge.min &&
				val.X < edge.max {
				return o
			}
		}
	}
	return s.orientation
}

func (s *state) update(val accelerometer.Value) {
	o := s.detectOrientation(val)
	if o != s.orientation {
		if s.newOrientation == nil || *s.newOrientation != o {
			s.newOrientation = &o
			s.ticks = 0
		} else {
			s.ticks++
			if s.ticks > s.autorotate.maxTicks {
				s.orientation = o
				s.newOrientation = nil
				s.ticks = 0
				err := s.autorotate.SetOrientation(o)
				if err != nil {
					fmt.Printf("Error changing orientation: %s\n", err.Error())
				}
			}
		}
	}
}
