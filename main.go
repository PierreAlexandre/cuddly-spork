package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Default settings
const (
	defaultConsulPort  = 8500
	defaultUpdateDelay = 1 * time.Second
	metricsFile        = "/tmp/node-exporter/tcp_connections.prom" // Prometheus textfile directory
)

var (
	consulPort  int
	updateDelay time.Duration
)

func init() {
	// Use environment variable for port if available
	if portStr, exists := os.LookupEnv("CONSUL_PORT"); exists {
		port, err := strconv.Atoi(portStr)
		if err == nil {
			consulPort = port
		} else {
			log.Printf("Invalid CONSUL_PORT value, using default: %d", defaultConsulPort)
			consulPort = defaultConsulPort
		}
	} else {
		consulPort = defaultConsulPort
	}

	// Use environment variable for update delay if available
	if delayStr, exists := os.LookupEnv("UPDATE_DELAY"); exists {
		delay, err := time.ParseDuration(delayStr)
		if err == nil {
			updateDelay = delay
		} else {
			log.Printf("Invalid UPDATE_DELAY value, using default: %s", defaultUpdateDelay)
			updateDelay = defaultUpdateDelay
		}
	} else {
		updateDelay = defaultUpdateDelay
	}

	log.Printf("Starting Prometheus TCP connection exporter:")
	log.Printf(" - Monitoring Port: %d", consulPort)
	log.Printf(" - Update Interval: %s", updateDelay)
}

func main() {
	ticker := time.NewTicker(updateDelay)
	defer ticker.Stop()

	for {
		ipv4Count := countOpenConnections("netstat -tan", consulPort)
		ipv6Count := countOpenConnections("netstat -tanp | grep tcp6", consulPort)

		log.Printf("Open IPv4 connections to port %d: %d", consulPort, ipv4Count)
		log.Printf("Open IPv6 connections to port %d: %d", consulPort, ipv6Count)

		writeMetrics(ipv4Count, ipv6Count)
		<-ticker.C
	}
}

// countOpenConnections runs `netstat` and counts active ESTABLISHED connections for both IPv4 and IPv6.
func countOpenConnections(command string, port int) int {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing command (%s): %v", command, err)
		return -1
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		state := fields[len(fields)-1] // Last field is the connection state
		localAddr := fields[3]         // Local Address (e.g., 127.0.0.1:8500 or [::]:8500)

		// Extract port number from local address
		extractedPort, err := extractPort(localAddr)
		if err == nil && extractedPort == port && state == "ESTABLISHED" {
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading command output: %v", err)
	}

	return count
}

// extractPort extracts the port number from an IPv4 or IPv6 address string.
func extractPort(address string) (int, error) {
	if strings.HasPrefix(address, "[") { // IPv6 format: [::1]:8500
		parts := strings.Split(strings.Trim(address, "[]"), "]:")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid IPv6 address format: %s", address)
		}
		return strconv.Atoi(parts[1])
	}

	parts := strings.Split(address, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid IPv4 address format: %s", address)
	}

	portStr := parts[len(parts)-1]
	return strconv.Atoi(portStr)
}

// writeMetrics writes the TCP connection counts to the Prometheus metrics file.
func writeMetrics(ipv4Count, ipv6Count int) {
	file, err := os.Create(metricsFile)
	if err != nil {
		log.Printf("Error writing to metrics file: %v", err)
		return
	}
	defer file.Close()

	metricContent := fmt.Sprintf("# HELP tcp_connections Number of open TCP connections on port %d\n", consulPort)
	metricContent += "# TYPE tcp_connections gauge\n"
	metricContent += fmt.Sprintf("tcp_connections{port=\"%d\", protocol=\"ipv4\"} %d\n", consulPort, ipv4Count)
	metricContent += fmt.Sprintf("tcp_connections{port=\"%d\", protocol=\"ipv6\"} %d\n", consulPort, ipv6Count)

	_, err = file.WriteString(metricContent)
	if err != nil {
		log.Printf("Error writing metric content: %v", err)
	}
}
