# Mikapod Poller
[![Go Report Card](https://goreportcard.com/badge/github.com/mikaponics/mikapod-poller)](https://goreportcard.com/report/github.com/mikaponics/mikapod-poller)

## Overview

The purpose of this application is to poll time-series data from our [Mikapod Soil Reader](https://github.com/mikaponics/mikapod-soil-reader) application and save it to the [Mikapod Storage](https://github.com/mikaponics/mikapod-storage) application. The interval of time is every one minute.

## Prerequisites

You must have the following installed before proceeding. If you are missing any one of these then you cannot begin.

* ``Go 1.12.7``

## Installation

1. Please visit the [Mikapod Soil (Arduino) device](https://github.com/mikaponics/mikapod-soil-arduino) repository and setup the external device and connect it to your development machine.

2. Please visit the [Mikapod Soil Reader](https://github.com/mikaponics/mikapod-soil-reader) repository and setup that application on your device.

3. Please visit the [Mikapod Storage](https://github.com/mikaponics/mikapod-storage) repository and setup that application on your device.

4. Get our latest code.

    ```
    go get -u github.com/mikaponics/mikapod-poller
    ```

5. Install the depencies for this project.

    ```
    go get -u google.golang.org/grpc
    go get -u github.com/tarm/serial
    ```

6. Run our application.

    ```
    cd github.com/mikaponics/mikapod-poller
    go run cmd/mikapod-poller/main.go
    ```


## License

This application is licensed under the **BSD** license. See [LICENSE](LICENSE) for more information.
