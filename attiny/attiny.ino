#include <TinyWireS.h>
#include <NewPing.h>

#define SONAR_TRIGGER_PIN 5
#define SONAR_ECHO_PIN 4

#define SONAR_MAX_DISTANCE 250 // cm

#define I2C_SLAVE_ADDRESS 0x4 // Address of the slave

// Operation to execute, sent from rpi
byte receive_op;

// NewPing setup of pins and maximum distance.
NewPing sonar(SONAR_TRIGGER_PIN, SONAR_ECHO_PIN, SONAR_MAX_DISTANCE);

// In any case, cannot be > SONAR_MAX_DISTANCE
byte sonar_distance;

/*
   Operations:

   - Turn LED on/off with I2C
*/

void setup()
{
  TinyWireS.begin(I2C_SLAVE_ADDRESS); // join i2c network

  TinyWireS.onReceive(receiveEvent);
  TinyWireS.onRequest(requestEvent);

  // Turn on LED when program starts
  pinMode(1, OUTPUT);
  digitalWrite(1, HIGH);
}

bool sonar_run = false;

void loop()
{
  /**
     This is the only way we can detect stop condition (http://www.avrfreaks.net/index.php?name=PNphpBB2&file=viewtopic&p=984716&sid=82e9dc7299a8243b86cf7969dd41b5b5#984716)
     it needs to be called in a very tight loop in order not to miss any (REMINDER: Do *not* use delay() anywhere, use tws_delay() instead).
     It will call the function registered via TinyWireS.onReceive(); if there is data in the buffer on stop.
  */
  TinyWireS_stop_check();
  
  switch (receive_op) {
    case 0:
      digitalWrite(1, LOW);
      break;
    case 1:
      digitalWrite(1, HIGH);
      break;
    case 2:
      sonar_run = true;
      break;
  }

  // Force reset
  receive_op = -1;

  if (sonar_run) {
    sonar_run = false;
    sonar_distance = sonar.ping_cm();
  }
}

// Gets called when the ATtiny receives an i2c request
void requestEvent()
{
  // In any case, cannot be > SONAR_MAX_DISTANCE  
  TinyWireS.send(sonar_distance);
}

/**
   The I2C data received -handler

   This needs to complete before the next incoming transaction (start, data, restart/stop) on the bus does
   so be quick, set flags for long running tasks to be called from the mainloop instead of running them directly,
*/
void receiveEvent(uint8_t howMany) {
  if (howMany != 1)
  {
    // Sanity-check, we expect a simple enum-style op
    return;
  }

  receive_op = TinyWireS.receive();
}

//
//void setup() {
//  //Initialisation of digital PINs
//  pinMode(0, OUTPUT); //LED on Model B
//  pinMode(1, OUTPUT); //LED on Model A or Pro
//}
//void loop() {
//  digitalWrite(0, HIGH); //turns on LED
//  digitalWrite(1, HIGH);
//  delay(1000); //waits a second
//  digitalWrite(0, LOW); //turns off LED
//  digitalWrite(1, LOW);
//delay(1000);
//}
