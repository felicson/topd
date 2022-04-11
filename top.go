package topd

import (
	"context"
	"fmt"
	"net/http"

	"net"
	"os"
	"time"
)

//NotFound handler
func NotFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}

//Run start web server
func Run(ctx context.Context, deps Deps, done chan struct{}) error {

	config := deps.GetConfig()
	logger := deps.GetLogger()

	if _, err := os.Stat(config.Socket); err == nil {
		logger.Info("Trying to remove exist socket file")
		if err := os.Remove(config.Socket); err != nil {
			return fmt.Errorf("on remove socket file: %v", err)
		}
	}

	addr, err := net.ResolveUnixAddr("unix", config.Socket)
	if err != nil {
		return err
	}

	ln, err := net.ListenUnix("unix", addr)

	//tcpAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:8081")
	//ln, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		return fmt.Errorf("create socket: %v", err)
	}
	defer ln.Close()

	if err = os.Chmod(config.Socket, 0777); err != nil {
		return fmt.Errorf("chmod socket: %v", err)
	}

	errChannel := make(chan error, 1)

	web := Web{
		siteMap:        deps.GetSiteCollection(),
		sessionPerSite: deps.GetSessionPerSite(),
		historyWriter:  deps.GetHistoryWriter(),
		bots:           deps.GetBotChecker(),
		logger:         logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/top/", web.logHandler(web.ErrHandler(web.TopServer)))
	mux.HandleFunc("/", NotFound)

	server := http.Server{
		Handler: mux,
	}

	go func(ln net.Listener) {
		if err = server.Serve(ln); err != nil {
			errChannel <- fmt.Errorf("listenAndServe start: %v", err)
		}

	}(ln)

	select {
	case err := <-errChannel:
		return err
	case <-done:
		_ = os.Remove(config.Socket)
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %v", err)
	}
	return nil
}
