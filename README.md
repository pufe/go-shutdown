# go-shutdown

Go shutdown helper is a library to execute graceful application shutdown. It watches multiple
services and if anyone fails or is terminated, it executes everyone else's close function.

# Example Usage

```
shutdown.Manage(graceTime).
  Listener("webserver", webserver.Start, webserver.Shutdown).
  PingCloseService("postgres", postgresClient.Ping, postgresClient.Close).
  Listen()
```

shutdown can manage multiple `PingCloseService`. All function arguments should be of type `func () error`.