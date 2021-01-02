# Mikapod Poller
[![Go Report Card](https://goreportcard.com/badge/github.com/mikaponics/mikapod-poller)](https://goreportcard.com/report/github.com/mikaponics/mikapod-poller)

## Overview

The purpose of this application is to poll time-series data from our [Mikapod Soil Reader](https://github.com/mikaponics/mikapod-soil-reader) application and save it to the [Mikapod Storage](https://github.com/mikaponics/mikapod-storage) application. The interval of time is every one minute.

## Prerequisites

You must have the following installed before proceeding. If you are missing any one of these then you cannot begin.

* ``Go 1.15``

## Installation

1. Please visit the [Mikapod Soil (Arduino) device](https://github.com/mikaponics/mikapod-soil-arduino) repository and setup the external device and connect it to your development machine.

2. Please visit the [Mikapod Soil Reader](https://github.com/mikaponics/mikapod-soil-reader) repository and setup that application on your device.

3. Please visit the [Mikapod Storage](https://github.com/mikaponics/mikapod-storage) repository and setup that application on your device.

4. Download the source code, build and install the application.

    ```
    GO111MODULE=on go get -u github.com/mikaponics/mikapod-poller
    ```

5. Run our application.

    ```
    mikapod-poller
    ```

## License

This application is licensed under the **BSD** license. See [LICENSE](LICENSE) for more information.
