package gouno

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gouno-agent-demo/config"
	"gouno-agent-demo/internal/agent"
	"gouno-agent-demo/internal/billing"
	"gouno-agent-demo/internal/gateway"
	"gouno-agent-demo/internal/httpapi"
	"gouno-agent-demo/internal/provider"
	"gouno-agent-demo/middleware"
	"gouno-agent-demo/router"
	"gouno-agent-demo/utility"
)

var webCmd = &cobra.Command{
	Use: "web",
	Run: startWebServer,
}

func init() {
	webCmd.Flags().StringP("config_path", "c", "./config", "config file path")
	webCmd.Flags().StringP("address", "a", "0.0.0.0", "address to listen on")
	webCmd.Flags().StringP("port", "p", "8080", "port to listen on")
	webCmd.Flags().BoolP("debug", "d", false, "debug mode")
	webCmd.Flags().StringP("env", "e", "production", "env: development, test, production")
}

func startWebServer(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	configPath := cmd.Flag("config_path").Value.String()
	env := cmd.Flag("env").Value.String()

	configManager, err := config.NewConfigManager(cmd, configPath, env)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	globalConfig := configManager.Config()

	if globalConfig.WebServerConfig.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	loggerLevel := zap.NewAtomicLevelAt(zapcore.Level(globalConfig.LogConfig.Level))
	logger := utility.NewLogger(loggerLevel)

	logger.Sugar().Info("starting web server...")

	// init db
	// defaultDriver := config.GlobalConfig.DatabaseConfig.GetDefaultDriver()
	// if defaultDriver == nil {
	// 	log.Fatalf("default driver not found")
	// }

	engine := gin.New()
	engine.Use(
		gin.Logger(),
		httpapi.RequestID(),
		middleware.RecoveryMiddleware(),
		middleware.SecurityHeadersMiddleware(!globalConfig.WebServerConfig.Debug),
		middleware.IPRateLimitMiddleware(ctx, globalConfig.WebServerConfig.RateLimitPerMinute, globalConfig.WebServerConfig.RateLimitMaxVisitors),
	)
	providers := make(map[string]provider.Provider)
	for _, name := range []string{"openai", "anthropic"} {
		providerConfig, ok := globalConfig.GatewayConfig.Providers[name]
		if !ok || !providerConfig.Enabled {
			continue
		}
		client, err := provider.NewHTTPProvider(name, providerConfig, globalConfig.GatewayConfig.AllowedHosts, globalConfig.GatewayConfig.RequestTimeout)
		if err != nil {
			log.Fatalf("configure %s provider: %v", name, err)
		}
		providers[name] = client
	}
	if _, ok := providers[globalConfig.GatewayConfig.DefaultProvider]; !ok {
		log.Fatalf("default provider %q is not enabled", globalConfig.GatewayConfig.DefaultProvider)
	}
	agentHandler := httpapi.New(agent.New(providers, globalConfig.GatewayConfig.DefaultProvider), gateway.NewAuthorizer(globalConfig.GatewayConfig.APIKeys), globalConfig.GatewayConfig, billing.NewLoggingUsageRecorder(logger), logger)
	router.RegisterWebRouter(engine, agentHandler)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", globalConfig.WebServerConfig.Address, globalConfig.WebServerConfig.Port),
		IdleTimeout:       globalConfig.WebServerConfig.IdleTimeout,
		WriteTimeout:      globalConfig.WebServerConfig.WriteTimeout,
		ReadTimeout:       globalConfig.WebServerConfig.ReadTimeout,
		ReadHeaderTimeout: globalConfig.WebServerConfig.ReadHeaderTimeout,
		Handler:           engine,
	}

	logger.Sugar().Infof("web server listening on %s", httpServer.Addr)
	logger.Sugar().Info("press Ctrl+C to exit")

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	logger.Sugar().Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	// Close

	logger.Sugar().Info("server exiting")
}
