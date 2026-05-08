package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

//go:embed gui_assets/*
var guiAssets embed.FS

func newGUICmd() *cobra.Command {
	var host string
	var openBrowser bool

	cmd := &cobra.Command{
		Use:   "gui",
		Short: "Open a modern local web GUI for Ollama",
		Long:  "Starts a local web interface with glassmorphism styling for common Ollama actions.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			listener, err := net.Listen("tcp", host)
			if err != nil {
				return err
			}
			defer listener.Close()

			sub, err := fs.Sub(guiAssets, "gui_assets")
			if err != nil {
				return err
			}

			mux := http.NewServeMux()
			mux.Handle("/", http.FileServer(http.FS(sub)))

			server := &http.Server{Handler: mux}
			url := fmt.Sprintf("http://%s", listener.Addr().String())

			fmt.Fprintf(os.Stdout, "Ollama GUI is running at %s\n", url)
			fmt.Fprintln(os.Stdout, "Press Ctrl+C to stop.")

			if openBrowser {
				_ = browser.OpenURL(url)
			}

			errCh := make(chan error, 1)
			go func() { errCh <- server.Serve(listener) }()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			select {
			case sig := <-sigCh:
				_ = sig
				_ = server.Close()
				return nil
			case err := <-errCh:
				if err == http.ErrServerClosed {
					return nil
				}
				return err
			}
		},
	}

	cmd.Flags().StringVar(&host, "host", "127.0.0.1:5173", "Host:port for the local GUI server")
	cmd.Flags().BoolVar(&openBrowser, "open", true, "Open the GUI in your browser")
	return cmd
}
