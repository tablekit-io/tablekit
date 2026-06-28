package cli

import (
	nethttp "net/http"

	"core/http"
	"core/services"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		appServices, err := services.New()
		if err != nil {
			return err
		}
		defer appServices.Close()
		app := http.Build(appServices)

		appServer := &nethttp.Server{Addr: ":" + appServices.Config.AppPort, Handler: app.AppEngine}
		controlServer := &nethttp.Server{Addr: ":" + appServices.Config.ControlPort, Handler: app.ControlEngine}

		g, ctx := errgroup.WithContext(cmd.Context())
		g.Go(func() error { return listen(appServer) })
		g.Go(func() error { return listen(controlServer) })
		g.Go(func() error {
			<-ctx.Done()
			_ = appServer.Close()
			_ = controlServer.Close()
			return nil
		})
		return g.Wait()
	},
}

// listen runs server, treating a graceful close as a clean exit.
func listen(server *nethttp.Server) error {
	if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
		return err
	}
	return nil
}
