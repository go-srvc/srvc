[![Go Reference](https://pkg.go.dev/badge/github.com/go-srvc/srvc.svg)](https://pkg.go.dev/github.com/go-srvc/srvc) [![codecov](https://codecov.io/github/go-srvc/srvc/graph/badge.svg?token=H3u7Ui9PfC)](https://codecov.io/github/go-srvc/srvc) ![main](https://github.com/go-srvc/srvc/actions/workflows/go.yaml/badge.svg?branch=main)

# Simple, Safe, and Modular Service Runner

srvc library provides a simple but powerful interface with zero external dependencies for running service modules.

## Use Case

Usually, Go services are composed of multiple "modules" which runs each in their own goroutine such as http server, signal listener, kafka consumer, ticker, etc. These modules should remain alive throughout the lifecycle of the whole service, and if one goes down, gracefully exit should be executed to avoid "zombie" services. srvc takes care of all this via a simple module interface.

## Acknowledgements

This library is something I have found myself writing over and over again in every project I been part of. One of the iterations can be found under [https://github.com/elisasre/go-common](https://github.com/elisasre/go-common).
