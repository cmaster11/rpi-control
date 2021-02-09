package main

import (
	"log"
	"time"

	"github.com/d2r2/go-i2c"
)

func main() {
	// Create a connection with attiny
	conn, err := i2c.New(0x4, 1)
	if err != nil {
		log.Fatal(err)
	}
	// Free I2C connection on exit
	defer conn.Close()

	go loopRead(conn)

	loop(conn)
}

func loopRead(conn *i2c.Options) {
	for {

		{
			// Read sonar value
			bytes := []byte{0, 0}
			numRead, err := conn.ReadBytes(bytes)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Read %d bytes\n", numRead)
			if len(bytes) > 0 {
				log.Printf("Value: %d", bytes[0])
			}
		}

		time.Sleep(250 * time.Millisecond)

	}
}

func loop(conn *i2c.Options) {

	for {

		{
			// Trigger sonar read
			_, err := conn.WriteBytes([]byte{0x2})
			if err != nil {
				log.Fatal(err)
			}
		}

		time.Sleep(500 * time.Millisecond)

	}
}
