/**
抄的，已应用到生产环境，如下是部分代码
来源：https://gist.github.com/akhenakh/38dbfea70dc36964e23acc19777f3869
 */
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
	"golang.org/x/sync/errgroup"
)

func main() {
	defer func() {
		fmt.Println("main")
		if err := recover(); err != nil {
			fmt.Println("panic")
			fmt.Println(err)
		}
	}()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	g, ctx := errgroup.WithContext(ctx)

	//http服务
	if config.Basic.App.PortOutHttp > 0 {
		g.Go(func() error {
			logPath := path.Join(config.Basic.App.LogDir, fmt.Sprintf("%s-%s.log", config.Basic.App.ServerName, "http"))
			f, _ := os.Create(logPath)
			gin.DefaultWriter = io.MultiWriter(f)
			if !config.Basic.App.Debug {
				gin.SetMode(gin.ReleaseMode)
			}
			app := routers.InitRouter()

			httpServer = &http.Server{
				Addr:    fmt.Sprintf(":%d", config.Basic.App.PortInnerHttp),
				Handler: app,
			}

			fmt.Println("启用http服务")
			return httpServer.ListenAndServe()
		})
	}

	//grpc服务
	if config.Basic.App.PortOutGrpc > 0 {
		g.Go(func() error {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Basic.App.PortInnerGrpc))
			if err != nil {
				return err
			}
			grpcServer = grpc.NewServer()
			pb.RegisterResoucePackServer(grpcServer, &server{})
			hsrv := health.NewServer()
			hsrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
			healthpb.RegisterHealthServer(grpcServer, hsrv)

			fmt.Println("启用grpc服务")

			return grpcServer.Serve(lis)
		})
	}

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	fmt.Println("received shutdown signal")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if httpServer != nil {
		_ = httpServer.Shutdown(shutdownCtx)
	}
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	err := g.Wait()
	if err != nil {
		fmt.Printf("server error: %v \n", err)
		return
	}
}
