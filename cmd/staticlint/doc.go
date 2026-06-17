// Package staticlint provides a multichecker for static analysis of Go code.
//
// # Overview
//
// staticlint is a custom static analysis tool built using the
// golang.org/x/tools/go/analysis/multichecker package. It combines multiple
// analyzers into a single binary to provide comprehensive static analysis
// for Go projects.
//
// # Usage
//
//	# Run on current package
//	staticlint .
//
//	# Run on all packages in the project
//	staticlint ./...
//
//	# Run with verbose output
//	staticlint -v ./...
//
//	# Get help
//	staticlint help
//
// # Included Analyzers
//
// ## Standard Analyzers (golang.org/x/tools/go/analysis/passes)
//
//   - printf: Checks printf-like formatting functions for correct usage
//   - shadow: Detects shadowed variables
//   - structtag: Validates struct field tags
//   - shift: Checks for bit shifts that may overflow
//   - lostcancel: Finds lost cancel functions returned by context.WithCancel
//   - errorsas: Validates errors.As calls
//   - assign: Detects useless assignments
//   - atomic: Checks for common mistakes in atomic operations
//   - defers: Check defer usage for common mistakes
//   - nilness: Detects nilness issues in code
//
// ## Staticcheck Analyzers (honnef.co/go/tools)
//
//   - SA-class: All SA analyzers (SA1000-SA9008) covering common bugs and
//     correctness issues. Includes checks like:
//
//   - SA1000: Invalid JSON calls
//
//   - SA1002: Invalid time.Parse format
//
//   - SA1003: Invalid comparison of os.ErrExist, os.ErrNotExist, os.ErrPermission
//
//   - SA1004: Comparing strings with comparison operators
//
//   - SA1005: Invalid call to time.Since
//
//   - SA1006: Printf with wrong argument type
//
//   - SA1007: Invalid URL in net/url.Parse
//
//   - SA1008: Non-canonical URL in net/url.Parse
//
//   - SA1010: Invalid regexp
//
//   - SA1011: Invalid call to strings.Replace
//
//   - SA1012: Invalid call to context.WithValue
//
//   - SA1013: Invalid call to io.Seek
//
//   - SA1014: Invalid call to os.Signal
//
//   - SA1015: Using time.Tick in defer
//
//   - SA1016: Invalid use of go/ast
//
//   - SA1017: Invalid call to os.Chmod
//
//   - SA1018: Using strings.Replace with zero repl argument
//
//   - SA1019: Using a deprecated function, variable, constant or field
//
//   - SA1020: Using an invalid argument to fmt.Sprintf
//
//   - SA1021: Invalid argument to sort.Slice
//
//   - SA1022: Invalid argument to sort.SliceStable
//
//   - SA1023: Invalid use of strconv
//
//   - SA1024: Invalid operation in select case
//
//   - SA2000: Test for error in deferred function
//
//   - SA2001: Test for error in go function
//
//   - SA2002: Test for error in defer of os.File.Close
//
//   - SA3000: Test for missing os.Exit
//
//   - SA4000: Invalid check for equality
//
//   - SA4001: Using & with struct literal
//
//   - SA4002: Test for switch with default
//
//   - SA4003: Invalid use of ClearOutputGroup
//
//   - SA4004: Test that the operand to panic is an error
//
//   - SA4005: Test that the operand to recover is not nil
//
//   - SA4006: A value assigned to a variable is never read
//
//   - SA4008: Test that a variable is not used before it is set
//
//   - SA4009: Test that a variable is not used after it is set
//
//   - SA4010: Test that a variable is not used as both LHS and RHS
//
//   - SA4011: Test that a break statement is not used in a switch
//
//   - SA4012: Test that a field is not compared to nil
//
//   - SA4013: Test that a function is not compared to nil
//
//   - SA4014: Test that an if statement is not used instead of switch
//
//   - SA4015: Test that a return statement is not used in a function that returns nothing
//
//   - SA4016: Test that a variable is not used after it is set
//
//   - SA4017: Test that a variable is not used before it is set
//
//   - SA4018: Test that a variable is not used as both LHS and RHS
//
//   - SA4019: Test that a variable is not used after it is set
//
//   - SA4020: Test that a variable is not used before it is set
//
//   - SA4021: Test that a variable is not used as both LHS and RHS
//
//   - SA5000: Test for empty select statement
//
//   - SA5001: Test that a defer statement is not used after a return
//
//   - SA5002: Test that a defer statement is not used after a recover
//
//   - SA5003: Test that a defer statement is not used after a panic
//
//   - SA5004: Test that a defer statement is not used outside a function
//
//   - SA5005: Test that a defer statement is not used in a loop
//
//   - SA5008: Test that a defer statement is not used in a switch
//
//   - SA5009: Test that a defer statement is not used in a select
//
//   - SA5010: Test that a defer statement is not used in a type switch
//
//   - SA5011: Test that a defer statement is not used after a return
//
//   - SA6000: Test for calling regexp.MatchString in a loop
//
//   - SA6001: Test for calling regexp.Compile in a loop
//
//   - SA6002: Test for calling sync.Pool.Put with a large object
//
//   - SA6003: Test for calling os.Open in a loop
//
//   - SA6005: Test for calling io.Copy in a loop
//
//   - SA9001: Test for deferring a method that returns a value
//
//   - SA9002: Test for using a variable before it is set
//
//   - SA9003: Test for using a variable after it is set
//
//   - SA9004: Test for using a variable as both LHS and RHS
//
//   - SA9005: Test for using a variable as both LHS and RHS
//
//   - SA9006: Test for using a variable as both LHS and RHS
//
//   - SA9007: Test for using a variable as both LHS and RHS
//
//   - SA9008: Test for using a variable as both LHS and RHS
//
//   - ST1000: Checks for missing package comment (stylecheck)
//
//   - QF1001: Suggests simplifying expression by using := (quickfix)
//
//   - S1000: Suggests simplifying code by using a single 'if' statement (simple)
//
// ## Public Analyzers
//
//   - asciicheck: Detects non-ASCII identifiers in Go source files.
//     Reports identifiers that contain non-ASCII characters, which can cause
//     issues with some tools and editors.
//
//   - ireturn: Checks return values for interfaces.
//     Detects functions that return interfaces instead of concrete types,
//     which can help enforce proper API design.
//
// ## Custom Analyzer
//
//   - osexitcheck: Prohibits direct calls to os.Exit in the main function
//     of the main package. Direct calls to os.Exit prevent deferred functions
//     from being executed, which can lead to resource leaks and other issues.
//     The analyzer inspects the AST to find function declarations named "main"
//     in packages named "main" and reports any direct call to os.Exit within
//     the body of that function.
//
// # Configuration
//
// The analyzers loaded from staticcheck and public packages are configured
// via the config.json file, which should be located in the same directory
// as the staticlint executable. The configuration file has the following format:
//
//	{
//	    "staticcheck": {
//	        "SA": true,
//	        "ST": ["ST1000"],
//	        "QF": ["QF1001"],
//	        "S": ["S1000"]
//	    },
//	    "public": [
//	        "asciicheck",
//	        "ireturn"
//	    ]
//	}
//
// # Exit Codes
//
//   - 0: No issues found
//   - 1: Issues were found
//   - 2: An error occurred during analysis
//
// # Building
//
// To build the staticlint binary:
//
//	go build -o staticlint ./cmd/staticlint/
//
// Then run:
//
//	./staticlint ./...
package main
