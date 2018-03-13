# Sample HTTP Application

**Note:** This repository should be imported as code.cloudfoundry.org/sample-http-app.

The sample http application implements the correct shutdown behavior for http applications deployed to Cloud Foundry. The sample http app adheres to the contract below when shutting down:

- App receives termination signal
- App closes listener so that it stops accepting new connections
- App finishes serving in-flight requests
- App closes existing connections as their requests complete
- App shuts down (or is KILLed)

Details of implementing the shutdown behavior can be found in [golang's http server Shutdown function](https://golang.org/src/net/http/server.go?s=78921:78975#L2552https://golang.org/src/net/http/server.go?s=78921:78975#L2552).

## Configure

The app is configurable by setting the following environment variables:

- `PORT`: The port  on which the HTTP server will be listening. It defaults to `8080`.
- `WAIT_TIME`: Wait time duration before the HTTP server responds to a request. Defaults to `1s`.

## Deploy to Cloud Foundry

Simply run `cf push` form the app root directory and it should deploy successfully.

## Run Locally

To run locally follow the steps below:

```
go get ...
go build .
./sample-http-app
```

To run the tests, run `ginkgo` in the app root directory:
