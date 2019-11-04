package accelerometer

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	Home = "/sys/bus/iio/devices"
)

type Value struct {
	X float64
	Y float64
}

func DetectAccelerometer() (string, error) {
	matches, err := filepath.Glob(Home + "/iio:device*/in_accel_x_raw")
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", errors.New("no accelerometers found")
	}
	return strings.Replace(strings.Replace(matches[0], Home+"/", "", 1), "/in_accel_x_raw", "", 1), nil
}

func NewReader(accelerometer string) *Reader {
	return &Reader{
		accelerometer: accelerometer,
	}
}
