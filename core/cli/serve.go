package cli

import (
	nethttp "net/http"

	"core/http"
	"core/services"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Msg("starting server")
		appServices, err := services.New()
		if err != nil {
			return err
		}
		defer appServices.Close()
		app := http.Build(appServices)

		appServer := &nethttp.Server{Addr: ":" + appServices.Config.AppPort, Handler: app.AppEngine}
		controlServer := &nethttp.Server{Addr: ":" + appServices.Config.ControlPort, Handler: app.ControlEngine}

		g, ctx := errgroup.WithContext(cmd.Context())
		g.Go(func() error {
			log.Info().Str("port", appServices.Config.AppPort).Msg("app server listening")
			return listen(appServer)
		})
		g.Go(func() error {
			log.Info().Str("port", appServices.Config.ControlPort).Msg("control server listening")
			return listen(controlServer)
		})
		g.Go(func() error { return appServices.WatchDatabasesFile(ctx) })
		g.Go(func() error {
			<-ctx.Done()
			log.Info().Msg("shutdown signal received, closing servers")
			if closeErr := appServer.Close(); closeErr != nil {
				log.Warn().Err(closeErr).Msg("app server close failed")
			}
			if closeErr := controlServer.Close(); closeErr != nil {
				log.Warn().Err(closeErr).Msg("control server close failed")
			}
			return nil
		})

		if waitErr := g.Wait(); waitErr != nil {
			log.Error().Err(waitErr).Msg("server exited")
			return waitErr
		}
		log.Info().Msg("server stopped")
		return nil
	},
}

// listen runs server, treating a graceful close as a clean exit.
func listen(server *nethttp.Server) error {
	if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
		log.Error().Str("addr", server.Addr).Err(err).Msg("http server failed")
		return err
	}
	return nil
}
