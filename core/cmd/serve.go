package cmd

import (
	"net/http"

	"core/config"
	"core/server"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		app, err := server.Build(cfg)
		if err != nil {
			return err
		}

		appSrv := &http.Server{Addr: ":" + cfg.AppPort, Handler: app.AppEng}
		controlSrv := &http.Server{Addr: ":" + cfg.ControlPort, Handler: app.Control}

		g, ctx := errgroup.WithContext(cmd.Context())
		g.Go(func() error { return listen(appSrv) })
		g.Go(func() error { return listen(controlSrv) })
		g.Go(func() error {
			<-ctx.Done()
			_ = appSrv.Close()
			_ = controlSrv.Close()
			return nil
		})
		return g.Wait()
	},
}

// listen runs srv, treating a graceful close as a clean exit.
func listen(srv *http.Server) error {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
