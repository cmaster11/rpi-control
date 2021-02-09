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

	loop(conn)
}

func loop(conn *i2c.Options) {
	on := false

	for {

		// Alternate LED

		if !on {
			log.Println("Turning ON")

			_, err := conn.WriteBytes([]byte{0x1})
			if err != nil {
				log.Fatal(err)
			}

			on = true
		} else {
			log.Println("Turning OFF")

			_, err := conn.WriteBytes([]byte{0x0})
			if err != nil {
				log.Fatal(err)
			}

			on = false
		}

		time.Sleep(2 * time.Second)

	}
}
