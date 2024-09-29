package services

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var envConfig = make(map[string]string)

func LoadConfig() {
	env := flag.String("env", "prod", "Specify the environment (dev or prod)")
	flag.Parse()
	log.Printf("Starting application with %s configuration", *env)

	filename := fmt.Sprintf("config-%s.properties", *env)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("could not open config file %s: %s", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envConfig[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading config file %s: %s", filename, err)
	}
}

func GetConfig(propertyName string) string {
	value, ok := envConfig[propertyName]
	if !ok {
		log.Fatalf("%s not found in config", propertyName)
	}
	return value
}
