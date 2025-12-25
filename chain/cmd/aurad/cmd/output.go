// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
	OutputFormatYAML OutputFormat = "yaml"
	OutputFormatCSV  OutputFormat = "csv"
)

// OutputFormatter handles different output formats
type OutputFormatter struct {
	format  OutputFormat
	writer  io.Writer
	verbose bool
	debug   bool
}

// NewOutputFormatter creates a new output formatter
func NewOutputFormatter(format string, verbose, debug bool) *OutputFormatter {
	return &OutputFormatter{
		format:  OutputFormat(format),
		writer:  os.Stdout,
		verbose: verbose,
		debug:   debug,
	}
}

// SetWriter sets the output writer
func (f *OutputFormatter) SetWriter(w io.Writer) {
	f.writer = w
}

// Print formats and prints data based on the configured format
func (f *OutputFormatter) Print(data interface{}) error {
	switch f.format {
	case OutputFormatJSON:
		return f.printJSON(data)
	case OutputFormatYAML:
		return f.printYAML(data)
	case OutputFormatCSV:
		return f.printCSV(data)
	default:
		return f.printText(data)
	}
}

// printJSON prints data as JSON
func (f *OutputFormatter) printJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	if f.verbose {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// printYAML prints data as YAML
func (f *OutputFormatter) printYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

// printCSV prints data as CSV
func (f *OutputFormatter) printCSV(data interface{}) error {
	writer := csv.NewWriter(f.writer)
	defer writer.Flush()

	// Handle different data types
	switch v := data.(type) {
	case []map[string]interface{}:
		return f.printMapSliceCSV(writer, v)
	case []interface{}:
		return f.printInterfaceSliceCSV(writer, v)
	case map[string]interface{}:
		return f.printMapCSV(writer, v)
	default:
		return fmt.Errorf("unsupported data type for CSV output: %T", data)
	}
}

// printMapSliceCSV prints a slice of maps as CSV
func (f *OutputFormatter) printMapSliceCSV(writer *csv.Writer, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Get headers from first item
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	// Write headers
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write rows
	for _, row := range data {
		var values []string
		for _, header := range headers {
			values = append(values, fmt.Sprintf("%v", row[header]))
		}
		if err := writer.Write(values); err != nil {
			return err
		}
	}

	return nil
}

// printInterfaceSliceCSV prints a slice of interfaces as CSV
func (f *OutputFormatter) printInterfaceSliceCSV(writer *csv.Writer, data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Try to extract headers from first item
	firstItem := data[0]
	val := reflect.ValueOf(firstItem)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		// Simple slice - just write values
		for _, item := range data {
			if err := writer.Write([]string{fmt.Sprintf("%v", item)}); err != nil {
				return err
			}
		}
		return nil
	}

	// Extract field names as headers
	typ := val.Type()
	var headers []string
	for i := 0; i < val.NumField(); i++ {
		headers = append(headers, typ.Field(i).Name)
	}

	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data rows
	for _, item := range data {
		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		var values []string
		for i := 0; i < val.NumField(); i++ {
			values = append(values, fmt.Sprintf("%v", val.Field(i).Interface()))
		}

		if err := writer.Write(values); err != nil {
			return err
		}
	}

	return nil
}

// printMapCSV prints a single map as CSV
func (f *OutputFormatter) printMapCSV(writer *csv.Writer, data map[string]interface{}) error {
	// Write headers
	var headers []string
	var values []string

	for key, value := range data {
		headers = append(headers, key)
		values = append(values, fmt.Sprintf("%v", value))
	}

	if err := writer.Write(headers); err != nil {
		return err
	}

	return writer.Write(values)
}

// printText prints data as formatted text
func (f *OutputFormatter) printText(data interface{}) error {
	switch v := data.(type) {
	case string:
		fmt.Fprintln(f.writer, v)
	case map[string]interface{}:
		f.printMap(v, 0)
	case []interface{}:
		for i, item := range v {
			if f.verbose {
				fmt.Fprintf(f.writer, "[%d] ", i)
			}
			f.printText(item)
		}
	default:
		fmt.Fprintf(f.writer, "%v\n", v)
	}
	return nil
}

// printMap prints a map with indentation
func (f *OutputFormatter) printMap(m map[string]interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	for key, value := range m {
		switch v := value.(type) {
		case map[string]interface{}:
			fmt.Fprintf(f.writer, "%s%s:\n", prefix, key)
			f.printMap(v, indent+1)
		case []interface{}:
			fmt.Fprintf(f.writer, "%s%s:\n", prefix, key)
			for i, item := range v {
				if m2, ok := item.(map[string]interface{}); ok {
					fmt.Fprintf(f.writer, "%s  [%d]:\n", prefix, i)
					f.printMap(m2, indent+2)
				} else {
					fmt.Fprintf(f.writer, "%s  [%d]: %v\n", prefix, i, item)
				}
			}
		default:
			fmt.Fprintf(f.writer, "%s%s: %v\n", prefix, key, value)
		}
	}
}

// Debug prints debug information if debug mode is enabled
func (f *OutputFormatter) Debug(format string, args ...interface{}) {
	if f.debug {
		fmt.Fprintf(f.writer, "[DEBUG] "+format+"\n", args...)
	}
}

// Verbose prints verbose information if verbose mode is enabled
func (f *OutputFormatter) Verbose(format string, args ...interface{}) {
	if f.verbose {
		fmt.Fprintf(f.writer, "[VERBOSE] "+format+"\n", args...)
	}
}

// Error prints error messages
func (f *OutputFormatter) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

// Warning prints warning messages
func (f *OutputFormatter) Warning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[WARNING] "+format+"\n", args...)
}

// Info prints informational messages
func (f *OutputFormatter) Info(format string, args ...interface{}) {
	fmt.Fprintf(f.writer, "[INFO] "+format+"\n", args...)
}
