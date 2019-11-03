package autorotate

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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

func OrientationMap() map[Orientation]LibInputCoordinates {
	return orientationMap
}

func SetOrientation(devices []string, display string, orientation Orientation) error {
	command, args := getXrandrCommand(orientation, display)
	fmt.Printf("%s %s\n", command, strings.Join(args, " "))
	err := executeCommand(command, args)
	if err != nil {
		return err
	}
	for _, device := range devices {
		command, args := getLibinputCommand(orientation, device)
		fmt.Printf("%s %s\n", command, strings.Join(args, " "))
		err := executeCommand(command, args)
		if err != nil {
			return err
		}
	}
	return nil
}

func getXrandrCommand(orientation Orientation, display string) (string, []string) {
	return "xrandr", []string{"-d", display, "-o", string(orientation)}
}

func getLibinputCommand(orientation Orientation, device string) (string, []string) {
	matrix := CoordinatesToString(orientationMap[orientation])
	args := []string{"set-prop", device, "Coordinate Transformation Matrix"}
	args = append(args, matrix...)
	return "xinput", args
}

func CoordinatesToString(coords LibInputCoordinates) []string {
	var strs []string
	for _, val := range coords {
		strs = append(strs, strconv.Itoa(val))
	}
	return strs
}

func executeCommand(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: false,
		Noctty:  false,
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
