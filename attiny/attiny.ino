#include <TinyWireS.h>
#include <NewPing.h>

#define SONAR_TRIGGER_PIN PB1
#define SONAR_ECHO_PIN PB3
#define SWITCH_UP_PIN PB4
#define SWITCH_DOWN_PIN PB5

#define SONAR_MAX_DISTANCE 250 // cm

#define MAX_SWITCH_ON_MILLIS 3000

#define I2C_SLAVE_ADDRESS 0x4 // Address of the slave

enum op {
  NOOP = 0,
  SONAR_RUN,
  SWITCH_ON,
  SWITCH_OFF,
  SWITCH_UP,
  SWITCH_DOWN
};

// Operation to execute, sent from rpi
op receive_op;

// NewPing setup of pins and maximum distance.
NewPing sonar(SONAR_TRIGGER_PIN, SONAR_ECHO_PIN, SONAR_MAX_DISTANCE);

// --- Local vars

// In any case, cannot be > SONAR_MAX_DISTANCE
byte sonar_distance = 0;
bool sonar_run = false;
bool switch_on = false;
bool switch_up = false;

unsigned long start_millis;

/*
   Operations:

   - Turn LED on/off with I2C
*/

void setup()
{
  pinMode(SONAR_TRIGGER_PIN, OUTPUT);
  pinMode(SONAR_ECHO_PIN, INPUT);
  pinMode(SWITCH_UP_PIN, OUTPUT);
  pinMode(SWITCH_DOWN_PIN, OUTPUT);

  digitalWrite(SONAR_TRIGGER_PIN, LOW);
  digitalWrite(SWITCH_UP_PIN, LOW);
  digitalWrite(SWITCH_DOWN_PIN, LOW);

  TinyWireS.begin(I2C_SLAVE_ADDRESS); // join i2c network

  TinyWireS.onReceive(receiveEvent);
  TinyWireS.onRequest(requestEvent);
}

void loop()
{
  /**
     This is the only way we can detect stop condition (http://www.avrfreaks.net/index.php?name=PNphpBB2&file=viewtopic&p=984716&sid=82e9dc7299a8243b86cf7969dd41b5b5#984716)
     it needs to be called in a very tight loop in order not to miss any (REMINDER: Do *not* use delay() anywhere, use tws_delay() instead).
     It will call the function registered via TinyWireS.onReceive(); if there is data in the buffer on stop.
  */
  TinyWireS_stop_check();

  switch (receive_op) {
    case SONAR_RUN:
      sonar_run = true;
      break;
    case SWITCH_ON:
      start_millis = millis();
      switch_on = true;
      break;
    case SWITCH_OFF:
      switch_on = false;
      break;
    case SWITCH_UP:
      switch_up = true;
      break;
    case SWITCH_DOWN:
      switch_up = false;
      break;
  }

  // Force reset
  receive_op = NOOP;

  if (sonar_run) {
    sonar_run = false;
    sonar_distance = sonar.ping_cm();
  }

  if (switch_on) {
    unsigned long last_millis = millis();

    if (last_millis - start_millis > MAX_SWITCH_ON_MILLIS) {
      // For safety, the switch will be turned off after a while, if not asked to be on
      switch_on = false;
    }
  }

  if (switch_on) {
    if (switch_up) {
      digitalWrite(SWITCH_UP_PIN, HIGH);
      digitalWrite(SWITCH_DOWN_PIN, LOW);
    } else {
      digitalWrite(SWITCH_UP_PIN, LOW);
      digitalWrite(SWITCH_DOWN_PIN, HIGH);
    }
  } else {
    digitalWrite(SWITCH_UP_PIN, LOW);
    digitalWrite(SWITCH_DOWN_PIN, LOW);
  }
}

// Gets called when the ATtiny receives an i2c request
void requestEvent()
{
  // In any case, cannot be > SONAR_MAX_DISTANCE
  TinyWireS.send(sonar_distance);
  TinyWireS.send(switch_on);
  TinyWireS.send(switch_up);
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

  receive_op = (op)TinyWireS.receive();
}
