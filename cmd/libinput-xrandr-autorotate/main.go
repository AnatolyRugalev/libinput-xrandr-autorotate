package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/AnatolyRugalev/libinput-xrandr-autorotate/pkg/accelerometer"
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
	var accelerometerName = flag.String("accelerometer", "", "accelerometer to use. Leave empty for autodetection")
	var threshold = flag.Float64("threshold", 7.0, "threshold for orientation edge detection")
	var refreshRate = flag.Int("refresh-rate", 200, "refresh rate in milliseconds")
	var ticks = flag.Int("ticks", 3, "wait for this ticks amount before applying changes")
	flag.Parse()

	var devices []string
	var err error
	if *touchscreensStr == "" {
		devices, err = autorotate.DetectTouchScreens()
		if err != nil {
			fmt.Printf("Cannot autodetect touchscreens: %s\n", err.Error())
			os.Exit(1)
		}
	} else {
		devices = strings.Split(*touchscreensStr, ",")
	}
	if *accelerometerName == "" {
		*accelerometerName, err = accelerometer.DetectAccelerometer()
		if err != nil {
			fmt.Printf("Cannot autodetect accelerometer: %s\n", err.Error())
			os.Exit(1)
		}
	}
	auto := autorotate.NewAutorotate(*display, devices, *accelerometerName, *threshold, time.Millisecond*time.Duration(*refreshRate), *ticks)
	exit := make(chan os.Signal, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer wg.Done()
		err := auto.Watch(ctx)
		if err != nil {
			fmt.Printf("Error starting watcher: %s", err.Error())
			os.Exit(1)
		}
	}()
	signal.Notify(exit, syscall.SIGINT)
	select {
	case <-exit:
		cancel()
		break
	}
	wg.Wait()
}
