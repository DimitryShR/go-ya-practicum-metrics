// Command staticlint implements a multichecker - a custom static analysis tool
// that combines multiple analyzers into a single binary.
//
// It includes the following analyzers:
//   - Standard analyzers from golang.org/x/tools/go/analysis/passes
//     (printf, shadow, structtag, shift, lostcancel, errorsas, assign, atomic)
//   - All SA-class analyzers from honnef.co/go/tools/staticcheck
//   - Selected analyzers from other staticcheck classes:
//     ST1000 (stylecheck - missing package comment),
//     QF1001 (quickfix - simplify expressions),
//     S1000 (simple - simplify code)
//   - Public analyzers: asciicheck, ireturn
//   - Custom analyzer osexitcheck - prohibits direct os.Exit calls in main function
//     of the main package
//
// Configuration:
//
//	The set of staticcheck analyzers is configured via config.json file
//	located in the same directory as the executable.
//
// Usage:
//
//	staticlint [-flag] [package]
//
// Run 'staticlint help' for more details,
// or 'staticlint help name' for details and flags of a specific analyzer.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"github.com/butuzov/ireturn/analyzer"
	"github.com/tdakkota/asciicheck"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/DimitryShR/go-ya-practicum-metrics/cmd/staticlint/analyzers/osexitcheck"
)

// Config describes the structure of config.json file.
type Config struct {
	Staticcheck StaticcheckConfig `json:"staticcheck"`
	Public      []string          `json:"public"`
}

// StaticcheckConfig describes which staticcheck analyzers to enable.
type StaticcheckConfig struct {
	SA bool     `json:"SA"`
	ST []string `json:"ST"`
	QF []string `json:"QF"`
	S  []string `json:"S"`
}

const configFileName = "config.json"

func main() {
	var mychecks []*analysis.Analyzer

	// 1. Add standard analyzers from golang.org/x/tools/go/analysis/passes
	mychecks = append(mychecks,
		printf.Analyzer,     // Check printf-like formatting
		shadow.Analyzer,     // Check for shadowed variables
		structtag.Analyzer,  // Check struct tags
		shift.Analyzer,      // Check bit shifts
		lostcancel.Analyzer, // Check for lost cancel functions
		errorsas.Analyzer,   // Check errors.As
		assign.Analyzer,     // Check useless assignments
		atomic.Analyzer,     // Check atomic operations
		defers.Analyzer,     // Check defer usage
		nilness.Analyzer,    // Check for nilness issues
	)

	// 2. Load configuration
	cfg, err := loadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
	} else {
		// 3. Add staticcheck analyzers based on config
		mychecks = append(mychecks, getStaticcheckAnalyzers(cfg)...)
	}

	// 4. Add public analyzers
	mychecks = append(mychecks, getPublicAnalyzers(cfg)...)

	// 5. Add custom analyzers
	mychecks = append(mychecks, osexitcheck.Analyzer)

	// Run multichecker
	multichecker.Main(mychecks...)
}

// loadConfig reads and parses the configuration file.
func loadConfig() (*Config, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("get executable path: %w", err)
	}

	configPath := filepath.Join(filepath.Dir(execPath), configFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

// getStaticcheckAnalyzers returns analyzers from staticcheck based on configuration.
func getStaticcheckAnalyzers(cfg *Config) []*analysis.Analyzer {
	if cfg == nil {
		return nil
	}

	var analyzers []*analysis.Analyzer

	// Map of analyzer names to include
	include := make(map[string]bool)

	// Add all SA-class analyzers if enabled
	if cfg.Staticcheck.SA {
		for _, a := range staticcheck.Analyzers {
			include[a.Analyzer.Name] = true
		}
	}

	// Add selected ST-class analyzers (stylecheck)
	for _, name := range cfg.Staticcheck.ST {
		include[name] = true
	}

	// Add selected QF-class analyzers (quickfix)
	for _, name := range cfg.Staticcheck.QF {
		include[name] = true
	}

	// Add selected S-class analyzers (simple)
	for _, name := range cfg.Staticcheck.S {
		include[name] = true
	}

	// Collect analyzers from all staticcheck classes
	allAnalyzers := make([]*analysis.Analyzer, 0)
	for _, a := range staticcheck.Analyzers {
		if include[a.Analyzer.Name] {
			allAnalyzers = append(allAnalyzers, a.Analyzer)
		}
	}
	for _, a := range stylecheck.Analyzers {
		if include[a.Analyzer.Name] {
			allAnalyzers = append(allAnalyzers, a.Analyzer)
		}
	}
	for _, a := range simple.Analyzers {
		if include[a.Analyzer.Name] {
			allAnalyzers = append(allAnalyzers, a.Analyzer)
		}
	}
	for _, a := range quickfix.Analyzers {
		if include[a.Analyzer.Name] {
			allAnalyzers = append(allAnalyzers, a.Analyzer)
		}
	}

	analyzers = append(analyzers, allAnalyzers...)

	return analyzers
}

// getPublicAnalyzers returns public analyzers based on configuration.
func getPublicAnalyzers(cfg *Config) []*analysis.Analyzer {
	if cfg == nil {
		return nil
	}

	var analyzers []*analysis.Analyzer

	for _, name := range cfg.Public {
		switch name {
		case "asciicheck":
			analyzers = append(analyzers, asciicheck.NewAnalyzer())
		case "ireturn":
			analyzers = append(analyzers, analyzer.NewAnalyzer())
		}
	}

	return analyzers
}
