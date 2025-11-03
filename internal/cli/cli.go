package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// CLI command interface for consistent command patterns
type CLICommand interface {
	Name() string
	Description() string
	Usage() string
	Execute(args []string) error
}

// CLIConfig holds CLI configuration settings
type CLIConfig struct {
	OutputFormat string // "human" or "json"
	Verbose      bool
	Quiet        bool
	Debug        bool
}

// CLIManager manages CLI consistency patterns
type CLIManager struct {
	commands map[string]CLICommand
	config   CLIConfig
}

// NewCLIManager creates a new CLI manager
func NewCLIManager() *CLIManager {
	return &CLIManager{
		commands: make(map[string]CLICommand),
		config:   CLIConfig{OutputFormat: "human"},
	}
}

// RegisterCommand registers a CLI command
func (c *CLIManager) RegisterCommand(cmd CLICommand) {
	c.commands[cmd.Name()] = cmd
}

// ExecuteCommand executes a CLI command
func (c *CLIManager) ExecuteCommand(name string, args []string) error {
	cmd, exists := c.commands[name]
	if !exists {
		return fmt.Errorf("unknown command: %s", name)
	}

	return cmd.Execute(args)
}

// ListCommands lists all available commands
func (c *CLIManager) ListCommands() {
	names := make([]string, 0, len(c.commands))
	for name := range c.commands {
		names = append(names, name)
	}
	sort.Strings(names)

	c.PrintOutput("Available commands:")
	for _, name := range names {
		cmd := c.commands[name]
		c.PrintOutput(fmt.Sprintf("  %s - %s", name, cmd.Description()))
	}
}

// PrintOutput formats output according to configuration
func (c *CLIManager) PrintOutput(message string) {
	if c.config.Quiet {
		return
	}

	if c.config.OutputFormat == "json" {
		output := map[string]string{"message": message}
		data, _ := json.Marshal(output)
		fmt.Println(string(data))
	} else {
		fmt.Println(message)
	}
}

// FormatError formats errors consistently
func (c *CLIManager) FormatError(err error) {
	if c.config.OutputFormat == "json" {
		output := map[string]string{
			"error":   "true",
			"message": err.Error(),
		}
		data, _ := json.Marshal(output)
		fmt.Fprintln(os.Stderr, string(data))
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
	}
}

// FormatMeasurements formats measurement results
func (c *CLIManager) FormatMeasurements(measurements []Measurement) {
	if c.config.OutputFormat == "json" {
		data, _ := json.MarshalIndent(measurements, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("Measurement Results:")
		fmt.Println(strings.Repeat("=", 60))
		for _, m := range measurements {
			fmt.Printf("Target: %s\n", m.Target)
			fmt.Printf("  Latency: %v\n", m.Latency)
			fmt.Printf("  Packet Loss: %.1f%%\n", m.PacketLoss)
			fmt.Printf("  Timestamp: %s\n", m.Timestamp.Format("2006-01-02 15:04:05.000"))
			fmt.Println(strings.Repeat("-", 40))
		}
	}
}

// Measurement represents ICMP measurement data
type Measurement struct {
	Target     string        `json:"target"`
	Latency    time.Duration `json:"latency"`
	PacketLoss float64       `json:"packet_loss"`
	Timestamp  time.Time     `json:"timestamp"`
}

// NewMeasurement creates a new measurement
func NewMeasurement(target string, latency time.Duration, packetLoss float64) *Measurement {
	return &Measurement{
		Target:     target,
		Latency:    latency,
		PacketLoss: packetLoss,
		Timestamp:  time.Now(),
	}
}
