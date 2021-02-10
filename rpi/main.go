package main

import (
	"log"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/davecgh/go-spew/spew"
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
	// status := readTinyStatus()
	return nil
}

func executeCmdDown(cmd *cobra.Command, args []string) error {
	return nil
}

func executeCmdMarkHigh(cmd *cobra.Command, args []string) error {
	return nil
}

func executeCmdMarkLow(cmd *cobra.Command, args []string) error {
	return nil
}

func executeCmdDebug(cmd *cobra.Command, args []string) error {
	if err := sendOp(OpSonarRun); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	_, err := readTinyStatus()
	if err != nil {
		return err
	}
	return nil
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
