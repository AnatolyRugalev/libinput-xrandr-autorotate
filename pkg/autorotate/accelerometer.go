package autorotate

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	AccelerometerHome = "/sys/bus/iio/devices"
)

func GetAccelerometer() (string, error) {
	matches, err := filepath.Glob(AccelerometerHome + "/iio:device*/in_accel_x_raw")
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", errors.New("no accelerometers found")
	}
	return strings.Replace(strings.Replace(matches[0], AccelerometerHome+"/", "", 1), "/in_accel_x_raw", "", 1), nil
}

func OpenAccelerometerValueFile(accelerometer string, name string) (*os.File, error) {
	path := AccelerometerHome + "/" + accelerometer + "/in_accel_" + name
	return os.OpenFile(path, os.O_RDONLY, 0)
}

func readFloat(r io.ReadSeeker) (float64, error) {
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return 0.0, err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return 0.0, err
	}
	return strconv.ParseFloat(strings.Trim(string(b), "\n"), 64)
}

func ReadValues(accelerometer string, exit <-chan struct{}, refreshRate time.Duration) (ch chan value, err error) {
	ch = make(chan value, 1)
	var sf, xf, yf, zf *os.File
	if sf, err = OpenAccelerometerValueFile(accelerometer, "scale"); err != nil {
		return
	}
	scale, err := readFloat(sf)
	_ = sf.Close()
	if err != nil {
		return
	}
	if xf, err = OpenAccelerometerValueFile(accelerometer, "x_raw"); err != nil {
		return
	}
	if yf, err = OpenAccelerometerValueFile(accelerometer, "y_raw"); err != nil {
		return
	}
	if zf, err = OpenAccelerometerValueFile(accelerometer, "z_raw"); err != nil {
		return
	}
	go func() {
		defer close(ch)
		defer xf.Close()
		defer yf.Close()
		defer zf.Close()
		for {
			select {
			case <-exit:
				return
			default:
				x, err := readFloat(xf)
				if err != nil {
					fmt.Printf("Cannot read value: %s\n", err.Error())
					return
				}
				y, err := readFloat(yf)
				if err != nil {
					fmt.Printf("Cannot read value: %s\n", err.Error())
					return
				}
				z, err := readFloat(zf)
				if err != nil {
					fmt.Printf("Cannot read value: %s\n", err.Error())
					return
				}
				x *= scale
				y *= scale
				z *= scale
				ch <- value{
					x: x,
					y: y,
					z: z,
				}
				time.Sleep(refreshRate)
			}
		}
	}()
	return ch, nil
}
