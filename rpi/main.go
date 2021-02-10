package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
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

const markFileNamePrefix = "tableHeight"
const markFileNameHigh = markFileNamePrefix + "_high"
const markFileNameLow = markFileNamePrefix + "_low"

type TinyStatus struct {

	/*
	  TinyWireS.send(sonar_distance);
	  TinyWireS.send(switch_on);
	  TinyWireS.send(switch_up);
	*/

	sonarDistance int
	switchOn      bool
	switchUp      bool
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
}

func main() {
	var err error

	// Init i2c connection
	{
		// Create a connection with attiny
		conn, err = i2c.New(0x4, 1)
		if err != nil {
			log.Fatal(err)
		}
		// Free I2C connection on exit
		defer conn.Close()
	}

	// Run CLI
	{
		err := Execute()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func executeCmdUp(cmd *cobra.Command, args []string) error {
	maxHeight, err := readHeight(markFileNameHigh)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Table max height: %d\n", maxHeight)
	}

	for {
		status, err := runSonarAndReadTinyStatus()
		if err != nil {
			return err
		}

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
		fmt.Printf("Error on exit: %s\n", err.Error())
	}

	return nil
}

func executeCmdDown(cmd *cobra.Command, args []string) error {
	minHeight, err := readHeight(markFileNameLow)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Table min height: %d\n", minHeight)
	}

	for {
		status, err := runSonarAndReadTinyStatus()
		if err != nil {
			return err
		}

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
		fmt.Printf("Error on exit: %s\n", err.Error())
	}

	return nil
}

func executeCmdMarkHigh(cmd *cobra.Command, args []string) error {
	status, err := runSonarAndReadTinyStatus()
	if err != nil {
		return err
	}

	if err := writeHeight(markFileNameHigh, status.sonarDistance); err != nil {
		return err
	}

	return nil
}

func executeCmdMarkLow(cmd *cobra.Command, args []string) error {
	status, err := runSonarAndReadTinyStatus()
	if err != nil {
		return err
	}

	if err := writeHeight(markFileNameLow, status.sonarDistance); err != nil {
		return err
	}

	return nil
}

func executeCmdDebug(cmd *cobra.Command, args []string) error {
	_, err := runSonarAndReadTinyStatus()
	if err != nil {
		return err
	}

	maxHeight, err := readHeight(markFileNameHigh)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Table max height: %d\n", maxHeight)
	}

	minHeight, err := readHeight(markFileNameLow)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Table min height: %d\n", minHeight)
	}

	return nil
}

func runSonarAndReadTinyStatus() (*TinyStatus, error) {
	if err := sendOp(OpSonarRun); err != nil {
		return nil, err
	}

	time.Sleep(400 * time.Millisecond)

	status, err := readTinyStatus()
	if err != nil {
		return nil, err
	}

	return status, nil
}

func readTinyStatus() (*TinyStatus, error) {
	status := &TinyStatus{}

	// Read values
	bytes := []byte{0, 0, 0}
	_, err := conn.ReadBytes(bytes)
	if err != nil {
		return nil, err
	}

	status.sonarDistance = int(bytes[0])
	status.switchOn = bytes[1] > 0
	status.switchUp = bytes[2] > 0

	spew.Printf("Read tiny status: %v\n", status)

	return status, nil
}

func readHeight(fileName string) (int, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return 0, errors.Errorf("Table height file %s does not exist", fileName)
	}

	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(string(bytes), 10, 32)
	if err != nil {
		return 0, err
	}

	return int(value), nil
}

func writeHeight(fileName string, value int) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d", value))
	if err != nil {
		return err
	}

	fmt.Printf("Height for file %s set at %d\n", fileName, value)

	return nil
}

func sendOp(op Op) error {
	_, err := conn.WriteBytes([]byte{byte(op)})
	return err
}

// func loopRead(conn *i2c.Options) {
// 	for {
//
// 		{
// 			// Read sonar value
// 			bytes := []byte{0, 0}
// 			numRead, err := conn.ReadBytes(bytes)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
//
// 			log.Printf("Read %d bytes\n", numRead)
// 			if len(bytes) > 0 {
// 				log.Printf("Value: %d", bytes[0])
// 			}
// 		}
//
// 		time.Sleep(250 * time.Millisecond)
//
// 	}
// }
//
// func loop(conn *i2c.Options) {
//
// 	for {
//
// 		{
// 			// Trigger sonar read
// 			_, err := conn.WriteBytes([]byte{0x2})
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 		}
//
// 		time.Sleep(500 * time.Millisecond)
//
// 	}
// }
