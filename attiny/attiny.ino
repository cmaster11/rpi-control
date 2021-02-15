#include "./libs/TinyWireS/usiTwiSlave.c"
#include "./libs/TinyWireS/TinyWireS.cpp"
#include "./libs/NewPing/src/NewPing.cpp"

// Manual mapping to physical pins
#define SONAR_TRIGGER_PIN 4
#define SONAR_ECHO_PIN 3
#define SWITCH_UP_PIN 1
#define SWITCH_DOWN_PIN 5

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
op last_receive_op;

// NewPing setup of pins and maximum distance.
// NewPing sonar(SONAR_TRIGGER_PIN, SONAR_ECHO_PIN, SONAR_MAX_DISTANCE);

// --- Local vars

unsigned long sonar_distance = 0;
bool sonar_run = false;
bool switch_on = false;
bool switch_up = false;

unsigned long start_millis;
unsigned long sonar_max_echo_time;
unsigned long sonar_max_time;

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

  // The relays are connected on low
  digitalWrite(SWITCH_UP_PIN, HIGH);
  digitalWrite(SWITCH_DOWN_PIN, HIGH);

  // Sonar setup
  sonar_max_echo_time = min(SONAR_MAX_DISTANCE, MAX_SENSOR_DISTANCE) * US_ROUNDTRIP_CM + (US_ROUNDTRIP_CM / 2);

  // I2C
  TinyWireS.begin(I2C_SLAVE_ADDRESS); // join i2c network
  TinyWireS.onReceive(receiveEvent);
  TinyWireS.onRequest(requestEvent);
}

void loop()
{
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
    sonar_distance = sonarPing();
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
      digitalWrite(SWITCH_UP_PIN, LOW);
      digitalWrite(SWITCH_DOWN_PIN, HIGH);
    } else {
      digitalWrite(SWITCH_UP_PIN, HIGH);
      digitalWrite(SWITCH_DOWN_PIN, LOW);
    }
  } else {
    digitalWrite(SWITCH_UP_PIN, HIGH);
    digitalWrite(SWITCH_DOWN_PIN, HIGH);
  }

  TinyWireS_stop_check();
}

bool sonarTriggerPing() {
  digitalWrite(SONAR_TRIGGER_PIN, LOW);   // Set the trigger pin low, should already be low, but this will make sure it is.
  delayMicroseconds(4);             // Wait for pin to go low.
  digitalWrite(SONAR_TRIGGER_PIN, HIGH);  // Set trigger pin high, this tells the sensor to send out a ping.
  delayMicroseconds(10);            // Wait long enough for the sensor to realize the trigger pin is high. Sensor specs say to wait 10uS.
  digitalWrite(SONAR_TRIGGER_PIN, LOW);   // Set trigger pin back to low.

  if (digitalRead(SONAR_ECHO_PIN)) {
    return false;     // Previous ping hasn't finished, abort.
  }

  sonar_max_time = micros() + sonar_max_echo_time + MAX_SENSOR_DELAY; // Maximum time we'll wait for ping to start (most sensors are <450uS, the SRF06 can take up to 34,300uS!)

  // Wait for ping to start.
  while (!digitalRead(SONAR_ECHO_PIN)) {
    // Took too long to start, abort.
    if (micros() > sonar_max_time) return false;
  }

  sonar_max_time = micros() + sonar_max_echo_time; // Ping started, set the time-out.
  return true;                         // Ping started successfully.
}

unsigned long sonarPing() {
  if (!sonarTriggerPing()) return NO_ECHO; // Trigger a ping, if it returns false, return NO_ECHO to the calling function.
  // Wait for the ping echo.
  while (digitalRead(SONAR_ECHO_PIN))  {
    if (micros() > sonar_max_time) return NO_ECHO; // Stop the loop and return NO_ECHO (false) if we're beyond the set maximum distance.
  }
  return (micros() - (sonar_max_time - sonar_max_echo_time) - PING_OVERHEAD); // Calculate ping time, include overhead.
}

// Gets called when the ATtiny receives an i2c request
void requestEvent()
{
  unsigned long to_send = sonar_distance;

  // unsigned int value = 0x1234;

  TinyWireS.send(to_send & 0xFF);
  TinyWireS.send((to_send >> 8) & 0xFF);
  TinyWireS.send((to_send >> 16) & 0xFF);
  TinyWireS.send((to_send >> 24) & 0xFF);

  TinyWireS.send(switch_on);
  TinyWireS.send(switch_up);
  TinyWireS.send(last_receive_op);
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
  last_receive_op = receive_op;
}
