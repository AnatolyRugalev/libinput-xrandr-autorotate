package accelerometer

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type Reader struct {
	accelerometer string
	scale         float64
	xf            *os.File
	yf            *os.File
}

func (r *Reader) openValueFile(name string) (*os.File, error) {
	path := Home + "/" + r.accelerometer + "/in_accel_" + name
	return os.OpenFile(path, os.O_RDONLY, 0)
}

func (r Reader) readFloat(f io.ReadSeeker) (float64, error) {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return 0.0, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return 0.0, err
	}
	return strconv.ParseFloat(strings.Trim(string(b), "\n"), 64)
}

func (r *Reader) Init() error {
	sf, err := r.openValueFile("scale")
	if err != nil {
		return err
	}
	r.scale, err = r.readFloat(sf)
	_ = sf.Close()
	if err != nil {
		return err
	}
	r.xf, err = r.openValueFile("x_raw")
	if err != nil {
		return err
	}
	r.yf, err = r.openValueFile("y_raw")
	if err != nil {
		return err
	}
	return nil
}

func (r *Reader) Read(ctx context.Context, refreshRate time.Duration, vals chan<- Value) {
	defer func() {
		close(vals)
		r.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Read: stop")
			return
		default:
			x, err := r.readFloat(r.xf)
			if err != nil {
				fmt.Printf("Cannot read value: %s\n", err.Error())
				return
			}
			y, err := r.readFloat(r.yf)
			if err != nil {
				fmt.Printf("Cannot read value: %s\n", err.Error())
				return
			}
			x *= r.scale
			y *= r.scale
			vals <- Value{
				X: x,
				Y: y,
			}
			time.Sleep(refreshRate)
		}
	}
}

func (r *Reader) Close() {
	_ = r.xf.Close()
	_ = r.yf.Close()
}
