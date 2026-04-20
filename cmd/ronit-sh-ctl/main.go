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
//	guestbook list                   list all guestbook entries (including hidden)
//	guestbook hide <id>              hide a guestbook entry
//	guestbook show <id>              unhide a guestbook entry
//	stats                            show aggregate stats
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ArkXero/termfolio/internal/storage"
)

var dbPath = flag.String("db", "/var/lib/termfolio/termfolio.db", "path to SQLite database")

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	db, err := storage.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch flag.Arg(0) {
	case "connections":
		cmdConnections(db)
	case "guestbook":
		if flag.NArg() < 2 {
			printUsage()
			os.Exit(1)
		}
		cmdGuestbook(db, flag.Args()[1:])
	case "stats":
		cmdStats(db)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", flag.Arg(0))
		printUsage()
		os.Exit(1)
	}
}

func cmdConnections(db *storage.DB) {
	sinceFlag := flag.String("since", "", "filter connections newer than this duration (e.g. 1h, 24h)")
	// Re-parse remaining args after "connections".
	flag.CommandLine.Parse(flag.Args()[1:]) //nolint:errcheck

	var since time.Time
	if *sinceFlag != "" {
		d, err := time.ParseDuration(*sinceFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid duration %q: %v\n", *sinceFlag, err)
			os.Exit(1)
		}
		since = time.Now().Add(-d)
	}

	records, err := db.ListConnections(since, 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(records) == 0 {
		fmt.Println("no connections found")
		return
	}

	fmt.Printf("%-4s  %-19s  %-19s  %-15s  %-18s  %-12s  %s\n",
		"ID", "started_at", "ended_at", "ip_prefix", "key_fp", "term", "size")
	fmt.Println(repeat("-", 110))
	for _, r := range records {
		ended := "-"
		if r.EndedAt != nil {
			ended = r.EndedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-4d  %-19s  %-19s  %-15s  %-18s  %-12s  %dx%d\n",
			r.ID,
			r.StartedAt.Format("2006-01-02 15:04:05"),
			ended,
			r.IPPrefix,
			r.KeyFP,
			r.Term,
			r.Width, r.Height,
		)
	}
}

func cmdGuestbook(db *storage.DB, args []string) {
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		entries, err := db.ListGuestbookEntries(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(entries) == 0 {
			fmt.Println("no entries")
			return
		}
		for _, e := range entries {
			hidden := ""
			if e.Hidden {
				hidden = " [HIDDEN]"
			}
			fmt.Printf("#%d  %s  %s%s\n  %s\n\n",
				e.ID,
				e.Name,
				e.CreatedAt.Format("2006-01-02 15:04"),
				hidden,
				e.Message,
			)
		}
	case "hide", "show":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "usage: guestbook %s <id>\n", args[0])
			os.Exit(1)
		}
		id, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid id %q\n", args[1])
			os.Exit(1)
		}
		hide := args[0] == "hide"
		if err := db.SetGuestbookEntryHidden(id, hide); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		action := "hidden"
		if !hide {
			action = "visible"
		}
		fmt.Printf("entry #%d marked %s\n", id, action)
	default:
		fmt.Fprintf(os.Stderr, "unknown guestbook subcommand: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func cmdStats(db *storage.DB) {
	stats, err := db.GetStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("total connections:    %d\n", stats.TotalConnections)
	fmt.Printf("unique visitors:      %d\n", stats.UniqueVisitors)
	if stats.MostPopularSection != "" {
		fmt.Printf("most popular section: %s\n", stats.MostPopularSection)
	} else {
		fmt.Printf("most popular section: (no data)\n")
	}
}

func printUsage() {
	fmt.Println("Usage: ronit-sh-ctl [--db path] <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  connections [--since duration]   list recent connections")
	fmt.Println("  guestbook list                   list all entries")
	fmt.Println("  guestbook hide <id>              hide an entry")
	fmt.Println("  guestbook show <id>              unhide an entry")
	fmt.Println("  stats                            aggregate stats")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --db path    SQLite database path (default: /var/lib/termfolio/termfolio.db)")
}

func repeat(s string, n int) string {
	r := ""
	for i := 0; i < n; i++ {
		r += s
	}
	return r
}
