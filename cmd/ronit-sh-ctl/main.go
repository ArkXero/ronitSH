// Command ronit-sh-ctl is the admin CLI for the ronit.sh SSH portfolio.
// It provides tools for moderating the guestbook and viewing connection stats.
//
// Usage:
//
//	ronit-sh-ctl [--db path] <command> [args]
//
// Commands:
//
//	connections [--since duration]   list recent SSH connections
//	guestbook list                   list all guestbook entries
//	guestbook hide <id>              hide a guestbook entry
//	guestbook show <id>              unhide a guestbook entry
//	stats                            show aggregate stats
package main

import (
	"flag"
	"fmt"
	"os"
)

var dbPath = flag.String("db", "/var/lib/termfolio/termfolio.db", "path to SQLite database")

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "connections":
		fmt.Println("connections subcommand -- coming in Phase 8")
	case "guestbook":
		fmt.Println("guestbook subcommand -- coming in Phase 8")
	case "stats":
		fmt.Println("stats subcommand -- coming in Phase 8")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", flag.Arg(0))
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: ronit-sh-ctl [--db path] <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  connections [--since duration]")
	fmt.Println("  guestbook list")
	fmt.Println("  guestbook hide <id>")
	fmt.Println("  guestbook show <id>")
	fmt.Println("  stats")
}
