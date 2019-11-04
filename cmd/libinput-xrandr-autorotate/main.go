package main

import (
	"flag"
	"fmt"
	"github.com/AnatolyRugalev/libinput-xrandr-autorotate/pkg/autorotate"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	var touchscreensStr = flag.String("touchscreens", "", "libinput touchscreen device names. Leave empty for autodetection")
	var display = flag.String("display", ":0", "xrandr displays")
	var accelerometer = flag.String("accelerometer", "", "accelerometer to use. Leave empty for autodetection")
	var threshold = flag.Float64("threshold", 7.0, "threshold for orientation edge detection")
	var refreshRate = flag.Int("refresh-rate", 200, "refresh rate in milliseconds")
	var ticks = flag.Int("ticks", 3, "wait for this ticks amount before applying changes")
	flag.Parse()

	var devices []string
	var err error
	if *touchscreensStr == "" {
		devices, err = autorotate.GetTouchScreens()
		if err != nil {
			fmt.Printf("Cannot autodetect touchscreens: %s\n", err.Error())
			os.Exit(1)
		}
	} else {
		devices = strings.Split(*touchscreensStr, ",")
	}
	if *accelerometer == "" {
		*accelerometer, err = autorotate.GetAccelerometer()
		if err != nil {
			fmt.Printf("Cannot autodetect accelerometer: %s\n", err.Error())
			os.Exit(1)
		}
	}
	exit := make(chan os.Signal, 1)
	stop := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := autorotate.Watch(stop, *display, devices, *accelerometer, *threshold, time.Millisecond*time.Duration(*refreshRate), *ticks)
		if err != nil {
			fmt.Printf("Error starting watcher: %s", err.Error())
			os.Exit(1)
		}
	}()
	signal.Notify(exit, syscall.SIGINT)
	select {
	case <-exit:
		close(stop)
		break
	}
	wg.Wait()
}
