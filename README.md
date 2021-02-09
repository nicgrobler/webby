
# webby

Simple http server in golang that features:  
- configurable timeouts
- optional signal handler for graceful shutdown

## Install

```golang
go get github.com/nicgrobler/webby
```

## Example usage  

```golang

// define a writehandler
func httpWriteHandler(w http.ResponseWriter, r *http.Request) {
	// handle CORS
	webby.SetCors(w, r)
  ...
}

// grab ctx to pass onto server(s)
ctx := webby.SignalContext()

// set a timeout for idle connections
idleTimeout := time.Duration(120*time.Second)

// set timeout that sets delay during shutdown before forcefully closing connections
forceTimeout := time.Duration(120*time.Second)

// initialise http listener
httpServer := webby.NewHTTPServer("http listener", "0.0.0.0:9090", idleTimeout)
httpServer.RegisterHandler("/route-of-your-choice", httpWriteHandler)

// run http server's listener in its own goroutine - with a "shutdown timeout"
go httpServer.StartListener(ctx, forceTimeout)

// wait for all to complete
<-httpServer.Done
```
