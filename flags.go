package main

// Verbose enables extra per-check logging for operator debugging.
var Verbose bool

// stripGlobalFlags removes -v/--verbose from args and returns the remainder.
func stripGlobalFlags(args []string) []string {
	Verbose = false
	var out []string
	for _, a := range args {
		switch a {
		case "-v", "--verbose":
			Verbose = true
		default:
			out = append(out, a)
		}
	}
	return out
}
