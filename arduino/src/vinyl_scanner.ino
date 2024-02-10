#include <Wire.h>
#include <Adafruit_PN532.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#include <HTTPClient.h>
#include <time.h>
#include <secrets.h>

#define TIMEZONE_OFFSET 1 * 3600
#define LED_BUILTIN     2

WiFiClientSecure client;
Adafruit_PN532 nfc(-1, &Serial2);

void setup() {
  Serial.begin(115200);
  pinMode(LED_BUILTIN, OUTPUT);
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

  // Configure time and notify
  configTime(TIMEZONE_OFFSET, 0, "pool.ntp.org");
}

String getTimestamp() {
  struct tm timeinfo;
  if(!getLocalTime(&timeinfo)){
    Serial.println("failed to obtain time");
    return String("");
  }

  char timestamp[20];
  strftime(timestamp, sizeof(timestamp), "%Y-%m-%d %H:%M:%S", &timeinfo);
  return String(timestamp);
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

void loop() {
  uint8_t success;
  uint8_t uid[] = { 0, 0, 0, 0, 0, 0, 0 };
  uint8_t uidLength;

  success = nfc.readPassiveTargetID(PN532_MIFARE_ISO14443A, uid, &uidLength); // Informações do NFC

  if (success) {
    String timestamp = getTimestamp();
    String uidString = formatUid(uidLength, uid);

    Serial.printf("%s: found an ISO14443A card\n", timestamp.c_str());
    Serial.printf("\tUID Length: %d bytes\n", uidLength);
    Serial.printf("\tUID Value: %s\n\n", uidString.c_str());

    digitalWrite(LED_BUILTIN, HIGH);
    sendVinylId(uidString);
    delay(1000);
    digitalWrite(LED_BUILTIN, LOW);
  }
}
