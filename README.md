# Sample HTTP Application

**Note:** This repository should be imported as `code.cloudfoundry.org/sample-http-app`.

The sample HTTP application implements the correct shutdown behavior for an HTTP application deployed to Cloud Foundry. It adheres to the following contract when shutting down:

- App receives termination signal
- App closes listener so that it stops accepting new connections
- App finishes serving in-flight requests
- App closes existing connections as their requests complete
- App shuts down (or is KILLed)

This application implements the connection management itself in [main.go](main.go), but similar logic can be found in the [Shutdown method of the Golang HTTP server](https://golang.org/src/net/http/server.go?s=78921:78975#L2552).


## Configure

The app is configurable by setting the following environment variables:

- `PORT`: The port on which the HTTP server will be listening. It defaults to `8080`.

## Deploy to Cloud Foundry

Simply run `cf push` from the app root directory and it should deploy successfully.

## Run Locally

To run locally, follow the steps below:

```
go get ...
go build .
./sample-http-app
```

To run the tests, run `ginkgo` in the app root directory:
