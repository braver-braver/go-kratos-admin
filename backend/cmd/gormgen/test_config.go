package main

import (
	"fmt"
	"log"
	"os"
)

// Test function to validate the configuration loading functionality
func testConfigLoading() {
	// Create a temporary test config file
	testConfig := `{
  "driver": "postgres",
  "dsn": "host=localhost user=gorm password=gorm dbname=gorm_test port=9920 sslmode=disable",
  "out": "test/output/path",
  "with_json_tag": false
}`

	// Write the test config to a file
	configFile := "test_config.json"
	err := os.WriteFile(configFile, []byte(testConfig), 0644)
	if err != nil {
		log.Fatalf("Failed to write test config file: %v", err)
	}
	defer os.Remove(configFile) // Clean up after test

	// Load the configuration
	var config Config
	err = loadConfig(configFile, &config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate the loaded configuration
	if config.Driver != "postgres" {
		log.Fatalf("Expected driver 'postgres', got '%s'", config.Driver)
	}
	if config.DSN != "host=localhost user=gorm password=gorm dbname=gorm_test port=9920 sslmode=disable" {
		log.Fatalf("DSN doesn't match expected value")
	}
	if config.Out != "test/output/path" {
		log.Fatalf("Expected output path 'test/output/path', got '%s'", config.Out)
	}
	if config.WithJSON != false {
		log.Fatalf("Expected with_json_tag false, got %t", config.WithJSON)
	}

	fmt.Println("✓ Configuration loading test passed!")
}
