package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Op int

const (
	OpSonarNoop Op = iota
	OpSonarRun
	OpSwitchOn
	OpSwitchOff
	OpSwitchUp
	OpSwitchDown
)

const sonarRoundTripUS = 57.0
const tableHeightTolerance = 0.4

const markFileNamePrefix = "tableHeight"
const markFileNameHigh = markFileNamePrefix + "_high"
const markFileNameLow = markFileNamePrefix + "_low"

const fileNameLock = markFileNamePrefix + "_lock"

type TinyStatus struct {

	/*
	  TinyWireS.send(sonar_distance);
	  TinyWireS.send(switch_on);
	  TinyWireS.send(switch_up);
	*/

	sonarDistance   float64
	sonarDistanceUS int64
	switchOn        bool
	switchUp        bool
	lastOperation   Op
}

var (
	rootCmd = &cobra.Command{
		Use: "rpi",
	}

	upCmd = &cobra.Command{
		Use:   "up",
		Short: "Moves the table up",
		RunE:  executeCmdUp,
	}

	downCmd = &cobra.Command{
		Use:   "down",
		Short: "Moves the table down",
		RunE:  executeCmdDown,
	}

	markHighCmd = &cobra.Command{
		Use:   "markHigh",
		Short: "Marks the current table height as max one",
		RunE:  executeCmdMarkHigh,
	}

	markLowCmd = &cobra.Command{
		Use:   "markLow",
		Short: "Marks the current table height as min one",
		RunE:  executeCmdMarkLow,
	}

	debug = &cobra.Command{
		Use:   "debug",
		Short: "Just prints current values",
		RunE:  executeCmdDebug,
	}

	printLoop = &cobra.Command{
		Use:   "printLoop",
		Short: "Prints distance in a loop until closed",
		RunE:  executeCmdPrintLoop,
	}

	printDistance = &cobra.Command{
		Use:   "printDistance",
		Short: "Just prints current distance",
		RunE:  executeCmdPrintDistance,
	}
)

var (
	conn *i2c.Options
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize()

	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(markHighCmd)
	rootCmd.AddCommand(markLowCmd)
	rootCmd.AddCommand(debug)
	rootCmd.AddCommand(printDistance)
	rootCmd.AddCommand(printLoop)
}

func main() {
	// Try to lock for a while
	locked := false
	for i := 0; i < 30; i++ {
		if err := lock(); err != nil {
			time.Sleep(150 * time.Millisecond)
			continue
		}

		locked = true
		break
	}

	if !locked {
		logrus.Fatal("Failed to lock")
	}
	defer unlock()

	logrus.SetLevel(logrus.DebugLevel)

	var err error

	// Init i2c connection
	{
		// Create a connection with attiny
		conn, err = i2c.New(0x4, 1)
		if err != nil {
			logrus.Error(err)
			return
		}
		// Free I2C connection on exit
		defer conn.Close()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			onExit()
		}
	}()

	// Run CLI
	{
		err := Execute()
		if err != nil {
			logrus.Error(err)
			return
		}
	}

	onExit()
}

func onExit() {
	if err := unlock(); err != nil {
		logrus.Error(err)
	}

	switchOff()

	os.Exit(0)
}

func lock() error {
	_, err := os.Stat(fileNameLock)
	if err != nil {
		if os.IsNotExist(err) {
			// NOOP
		} else {
			return err
		}
	} else {
		// If lock exists, error!
		return errors.Errorf("Lock file already exists")
	}

	{
		_, err := os.Create(fileNameLock)
		if err != nil {
			return err
		}
	}

	return nil
}

func unlock() error {
	_, err := os.Stat(fileNameLock)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		// If lock exists, remove!
		err := os.Remove(fileNameLock)
		if err != nil {
			return err
		}
	}

	return nil
}

func executeCmdUp(cmd *cobra.Command, args []string) error {
	defer switchOff()

	maxHeight, err := readHeight(markFileNameHigh)
	if err != nil {
		logrus.Debugf("Error: %s\n", err.Error())
		return err
	} else {
		logrus.Debugf("Table max height: %.1f\n", maxHeight)
	}

	for {
		status, err := runSonarAndReadTinyStatus()
		if err != nil {
			return err
		}

		logrus.Info(spew.Sprintf("Read tiny status: %v\n", status))

		if status.sonarDistance >= maxHeight {
			break
		}

		if err := sendOp(OpSwitchUp); err != nil {
			return err
		}
		if err := sendOp(OpSwitchOn); err != nil {
			return err
		}
	}

	if err := sendOp(OpSwitchOff); err != nil {
		return err
	}

	// Double check the position
	for {
		distance, err := runSonarAndReadDistanceAvg(5, false)
		if err != nil {
			return err
		}

		if distance > maxHeight+tableHeightTolerance {
			// Go down a bit
			if err := sendOp(OpSwitchDown); err != nil {
				return err
			}
			if err := sendOp(OpSwitchOn); err != nil {
				return err
			}

			time.Sleep(300 * time.Millisecond)

			if err := sendOp(OpSwitchOff); err != nil {
				return err
			}
		} else if distance < maxHeight-tableHeightTolerance {
			// Too low!
			// Go up a bit
			if err := sendOp(OpSwitchUp); err != nil {
				return err
			}
			if err := sendOp(OpSwitchOn); err != nil {
				return err
			}

			time.Sleep(300 * time.Millisecond)

			if err := sendOp(OpSwitchOff); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func executeCmdDown(cmd *cobra.Command, args []string) error {
	defer switchOff()

	minHeight, err := readHeight(markFileNameLow)
	if err != nil {
		logrus.Debugf("Error: %s\n", err.Error())
		return err
	} else {
		logrus.Debugf("Table min height: %.1f\n", minHeight)
	}

	for {
		status, err := runSonarAndReadTinyStatus()
		if err != nil {
			return err
		}

		logrus.Info(spew.Sprintf("Read tiny status: %v\n", status))

		if status.sonarDistance <= minHeight {
			break
		}

		if err := sendOp(OpSwitchDown); err != nil {
			return err
		}
		if err := sendOp(OpSwitchOn); err != nil {
			return err
		}
	}

	if err := sendOp(OpSwitchOff); err != nil {
		return err
	}

	// Double check the position
	for {
		distance, err := runSonarAndReadDistanceAvg(5, false)
		if err != nil {
			return err
		}

		if distance < minHeight-tableHeightTolerance {
			// Go up a bit
			if err := sendOp(OpSwitchUp); err != nil {
				return err
			}
			if err := sendOp(OpSwitchOn); err != nil {
				return err
			}

			time.Sleep(300 * time.Millisecond)

			if err := sendOp(OpSwitchOff); err != nil {
				return err
			}
		} else if distance > minHeight+tableHeightTolerance {
			// Too high
			// Go down a bit
			if err := sendOp(OpSwitchDown); err != nil {
				return err
			}
			if err := sendOp(OpSwitchOn); err != nil {
				return err
			}

			time.Sleep(300 * time.Millisecond)

			if err := sendOp(OpSwitchOff); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func switchOff() {
	if err := sendOp(OpSwitchOff); err != nil {
		logrus.Debugf("Error on switch off: %s\n", err.Error())
	}
}

func executeCmdMarkHigh(cmd *cobra.Command, args []string) error {
	distance, err := runSonarAndReadDistanceAvg(10, true)
	if err != nil {
		return err
	}

	if err := writeHeight(markFileNameHigh, distance); err != nil {
		return err
	}

	return nil
}

func executeCmdMarkLow(cmd *cobra.Command, args []string) error {
	distance, err := runSonarAndReadDistanceAvg(10, true)
	if err != nil {
		return err
	}

	if err := writeHeight(markFileNameLow, distance); err != nil {
		return err
	}

	return nil
}

func executeCmdDebug(cmd *cobra.Command, args []string) error {
	_, err := runSonarAndReadDistanceAvg(5, true)
	if err != nil {
		return err
	}

	maxHeight, err := readHeight(markFileNameHigh)
	if err != nil {
		logrus.Debugf("Error: %s\n", err.Error())
	} else {
		logrus.Debugf("Table max height: %.1f\n", maxHeight)
	}

	minHeight, err := readHeight(markFileNameLow)
	if err != nil {
		logrus.Debugf("Error: %s\n", err.Error())
	} else {
		logrus.Debugf("Table min height: %.1f\n", minHeight)
	}

	return nil
}

func executeCmdPrintLoop(cmd *cobra.Command, args []string) error {
	for {
		status, err := runSonarAndReadTinyStatus()
		if err != nil {
			return err
		}

		logrus.Debugf("Current distance: %.1f (uS: %d, hex: %#X)\n", status.sonarDistance, status.sonarDistanceUS, status.sonarDistanceUS)

		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

func executeCmdPrintDistance(cmd *cobra.Command, args []string) error {
	// Do not print other stuff
	logrus.SetLevel(logrus.FatalLevel)

	distance, err := runSonarAndReadDistanceAvg(5, false)
	if err != nil {
		return err
	}

	fmt.Printf("%.1f", distance)

	return nil
}

func runSonarAndReadTinyStatus() (*TinyStatus, error) {
	if err := sendOp(OpSonarRun); err != nil {
		return nil, err
	}

	time.Sleep(50 * time.Millisecond)

	status, err := readTinyStatus()
	if err != nil {
		return nil, err
	}

	return status, nil
}

func runSonarAndReadDistanceAvg(runs int, debug bool) (float64, error) {
	avg := 0.0

	// Make an average of all measurements
	for i := 0; i < runs; i++ {
		status, err := runSonarAndReadTinyStatus()
		if err != nil {
			return 0, err
		}

		if i == 0 {
			avg = status.sonarDistance
		} else {
			avg = ((avg * float64(i)) + status.sonarDistance) / float64(i+1)
		}

		if debug {
			logrus.Debug(spew.Sprintf("Read distance: %.1f, current avg: %.1f\n", status.sonarDistance, avg))
		}
	}

	logrus.Info(spew.Sprintf("Read distance avg: %.1f\n", avg))

	return avg, nil
}

func readTinyStatus() (*TinyStatus, error) {
	status := &TinyStatus{}

	// Read values
	bytes := []byte{0, 0, 0, 0, 0}
	_, err := conn.ReadBytes(bytes)
	if err != nil {
		return nil, err
	}

	byteIndexRest := 2

	logrus.Debug(spew.Sprintf("Read tiny status bytes: %v\n", bytes))

	distanceUS := int64(bytes[0]) + (int64(bytes[1]) << 8)

	status.sonarDistanceUS = distanceUS
	status.sonarDistance = float64(distanceUS) / sonarRoundTripUS

	status.switchOn = bytes[byteIndexRest] > 0
	status.switchUp = bytes[byteIndexRest+1] > 0
	status.lastOperation = Op(bytes[byteIndexRest+2])

	logrus.Debug(spew.Sprintf("Read tiny status: %v\n", status))

	return status, nil
}

func readHeight(fileName string) (float64, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return 0, errors.Errorf("Table height file %s does not exist", fileName)
	}

	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(string(bytes)), 64)
	if err != nil {
		return 0, err
	}

	return float64(value), nil
}

func writeHeight(fileName string, value float64) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%.1f", value))
	if err != nil {
		return err
	}

	logrus.Debugf("Height for file %s set at %.1f\n", fileName, value)

	return nil
}

func sendOp(op Op) error {
	_, err := conn.WriteBytes([]byte{byte(op)})
	return err
}
