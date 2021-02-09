#include <TinyWireS.h>

#define I2C_SLAVE_ADDRESS 0x4 // Address of the slave

byte receive_op;

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

void loop()
{
  switch (receive_op) {
    case 0:
      digitalWrite(1, LOW);
      break;
    case 1:
      digitalWrite(1, HIGH);
      break;
  }

  /**
     This is the only way we can detect stop condition (http://www.avrfreaks.net/index.php?name=PNphpBB2&file=viewtopic&p=984716&sid=82e9dc7299a8243b86cf7969dd41b5b5#984716)
     it needs to be called in a very tight loop in order not to miss any (REMINDER: Do *not* use delay() anywhere, use tws_delay() instead).
     It will call the function registered via TinyWireS.onReceive(); if there is data in the buffer on stop.
  */
  TinyWireS_stop_check();
}

int i = 0;
// Gets called when the ATtiny receives an i2c request
void requestEvent()
{
  TinyWireS.send(i);
  i++;
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
