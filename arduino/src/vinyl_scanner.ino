#include <PN532.h>
#include <PN532_HSU.h>
#include <Adafruit_NeoPixel.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#include <HTTPClient.h>
#include <secrets.h>

#define TIMEZONE_OFFSET 1 * 3600
#define LED_BUILTIN     2
#define PIN_WS2812B 15  // The ESP32 pin GPIO16 connected to WS2812B
#define NUM_PIXELS 30   // The number of LEDs (pixels) on WS2812B LED strip

Adafruit_NeoPixel ws2812b(NUM_PIXELS, PIN_WS2812B, NEO_GRB + NEO_KHZ800);
unsigned long pixelsInterval=10;  // time to wait
unsigned long rainbowPreviousMillis=0;
unsigned long rainbowCyclesPreviousMillis=0;
int rainbowCycleCycles = 0;
uint16_t currentPixel = 0;// what pixel are we operating on

WiFiClientSecure client;
PN532_HSU pn532hsu(Serial2);
PN532 nfc(pn532hsu);

uint64_t timestamp_millis;
String prev_uid;
bool vinyl_present = false;
bool lights_on = false;

void setup() {
  Serial.begin(115200);
  delay(100);

  nfc.begin();
  // Setup PN53x / NFC module
  uint32_t nfcVersion = nfc.getFirmwareVersion();
  if (! nfcVersion) {
    Serial.print("PN53x not found");
    while (1); // block forever
  }
  Serial.print("Found chip PN5"); Serial.println((nfcVersion >> 24) & 0xFF, HEX);
  Serial.print("Firmware ver. "); Serial.print((nfcVersion >> 16) & 0xFF, DEC);
  Serial.print('.'); Serial.println((nfcVersion >> 8) & 0xFF, DEC);
  nfc.SAMConfig();
  Serial.println("NFC module initialized");

  // Setup Wi-Fi
  WiFi.mode(WIFI_STA);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("");
  Serial.println("WiFi connected");

  // Initialize WS2812B strip object (REQUIRED)
  ws2812b.begin();
  timestamp_millis = millis();
}

void sendVinylId(String id) {
  HTTPClient http;
  http.begin(ENDPOINT);
  http.addHeader("Authorization", "Token " + String(ENDPOINT_TOKEN));

  int statusCode = http.POST(id);
  if (statusCode != 200) {
    Serial.printf("HTTP GET to %s failed with status code: %d\n", ENDPOINT, statusCode);
  }
  http.end();
}

String formatUid(uint8_t length, uint8_t uid[]) {
  String str;
  for (byte i = 0; i < length; i++) {
    // Convert each byte to a hexadecimal string representation
    String byteString = String(uid[i], HEX);
    // If the byte is less than 0x10, add a leading zero for consistency
    if (uid[i] < 0x10) {
      byteString = "0" + byteString;
    }
    // Concatenate the byte string to the UID string
    str += byteString;
  }
  return str;
}

// rainbow code from https://github.com/ndsh/neopixel-without-delay/tree/master?tab=readme-ov-file
void rainbow() {
  for(uint16_t i = 0; i< ws2812b.numPixels(); i++) {
      ws2812b.setPixelColor(i, Wheel(((i * 256 / ws2812b.numPixels()) + rainbowCycleCycles) & 255));
  }
  ws2812b.show();

  rainbowCycleCycles++;
  if(rainbowCycleCycles >= 256*5) rainbowCycleCycles = 0;
}

uint32_t Wheel(byte WheelPos) {
  WheelPos = 255 - WheelPos;
  if(WheelPos < 85) {
    return ws2812b.Color(255 - WheelPos * 3, 0, WheelPos * 3);
  }
  if(WheelPos < 170) {
    WheelPos -= 85;
    return ws2812b.Color(0, WheelPos * 3, 255 - WheelPos * 3);
  }
  WheelPos -= 170;
  return ws2812b.Color(WheelPos * 3, 255 - WheelPos * 3, 0);
}

void loop() {
  uint8_t success;
  uint8_t uid[] = { 0, 0, 0, 0, 0, 0, 0 };
  uint8_t uidLength;
  success = nfc.readPassiveTargetID(PN532_MIFARE_ISO14443A, uid, &uidLength); // Informações do NFC

  if (success) {
    vinyl_present = true;
    timestamp_millis = millis();
    String uidString = formatUid(uidLength, uid);

    if (prev_uid != uidString){
      Serial.printf("Found an ISO14443A card\n");
      Serial.printf("\tUID Length: %d bytes\n", uidLength);
      Serial.printf("\tUID Value: %s\n\n", uidString.c_str());
      Serial.printf("\tPrevious value: %s\n\n", prev_uid.c_str());

      sendVinylId(uidString);
      prev_uid = uidString;
    }
  }

  // If vinyl is present keep playing rainbow animation
  if (vinyl_present){
    if ((unsigned long)(millis() - rainbowPreviousMillis) >= pixelsInterval) {
      rainbowPreviousMillis = millis();
      rainbow();
    }
  }
  // If there is no tag scanned for 500 ms turn off the led strip and reset vinyl_present variable
  if (millis() - timestamp_millis > 500 && vinyl_present){
    vinyl_present = false;
    ws2812b.clear();  
    ws2812b.show();  
  }
}
