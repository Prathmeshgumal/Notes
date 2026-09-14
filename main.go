// Command nib is a local-first Markdown note taker: a terminal UI by
// default, with an optional web interface served from the same binary.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Prathmeshgumal/nib/internal/store"
	"github.com/Prathmeshgumal/nib/internal/tui"
	"github.com/Prathmeshgumal/nib/internal/web"
)

var version = "dev"

func main() {
	var (
		webOnly = flag.Bool("web", false, "run only the web UI (no terminal interface)")
		port    = flag.Int("port", 4321, "port for the web UI")
		dbPath  = flag.String("db", tui.DefaultPath(), "path to the notes database")
		showVer = flag.Bool("version", false, "print the version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Println("nib", version)
		return
	}

	// Notes written under the program's previous name follow it here.
	if *dbPath == tui.DefaultPath() {
		moved, err := store.AdoptLegacy(*dbPath, tui.LegacyPath())
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: could not bring your old notes across:", err)
		} else if moved {
			fmt.Fprintf(os.Stderr, "Your notes moved to %s — the old copy is still at %s\n",
				*dbPath, tui.LegacyPath())
		}
	}

	// Copy the database before touching it, so a bad day is always recoverable.
	if _, err := store.Snapshot(*dbPath); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not write a snapshot:", err)
	}

	st, err := store.Open(*dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer st.Close()

	// A brand-new database opens on a welcome note rather than an empty screen.
	if err := st.SeedIfEmpty(); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not add the welcome note:", err)
	}

	if *webOnly {
		runWeb(st, *port)
		return
	}

	if _, err := tea.NewProgram(tui.New(st), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func runWeb(st *store.Store, port int) {
	srv, err := web.New(st, port)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	srv.Start()
	fmt.Printf("\n  nib web UI at %s\n  Database: %s\n  Ctrl+C to stop\n\n", srv.URL, st.Path)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	_ = srv.Stop()
}
