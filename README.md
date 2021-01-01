# k8s-virtual-device-plugin

[![Docker Automated build][automated-image]][automated-link]
[![Build Status][build-image]][build-link]
[![Coverage Status][cov-image]][cov-link]

You can control number of pods per node by using this plugin without actually having any devices.

## Development

```bash
$ go mod download
```

## Build

```bash
$ make build
```

## Build Docker Image

```bash
$ make build-image
```

## Test in Minikube

You can test the device plugin in minikube.

```bash
$ make minikube-load
$ kubectl apply -f example
```


[automated-image]: https://img.shields.io/docker/cloud/automated/dtaniwaki/k8s-virtual-device-plugin.svg
[automated-link]:  https://hub.docker.com/r/dtaniwaki/k8s-virtual-device-plugin
[build-image]: https://travis-ci.com/dtaniwaki/k8s-virtual-device-plugin.svg
[build-link]:  http://travis-ci.com/dtaniwaki/k8s-virtual-device-plugin
[cov-image]:   https://coveralls.io/repos/github/dtaniwaki/k8s-virtual-device-plugin/badge.svg?branch=main
[cov-link]:    https://coveralls.io/github/dtaniwaki/k8s-virtual-device-plugin?branch=main
