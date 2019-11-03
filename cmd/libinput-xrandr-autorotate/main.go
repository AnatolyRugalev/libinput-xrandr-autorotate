package main

import (
	"flag"
	"fmt"
	"github.com/AnatolyRugalev/libinput-xrandr-autorotate/pkg/autorotate"
	"os"
	"strings"
)

func main() {
	var devices = flag.String("device", "Wacom HID 4875 Finger", "libinput touchscreen device names")
	var display = flag.String("display", ":0", "xrandr display number")
	var set = flag.String("set", "normal", "tmp")
	flag.Parse()
	if set != nil {
		err := autorotate.SetOrientation(strings.Split(*devices, ","), *display, autorotate.Orientation(*set))
		if err != nil {
			fmt.Printf("Error occured: %s", err.Error())
			os.Exit(1)
		}
	}
}
