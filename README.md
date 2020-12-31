# k8s-virtual-device-plugin

[![Docker Automated build](https://img.shields.io/docker/automated/dtaniwaki/k8s-virtual-device-plugin.svg)](https://hub.docker.com/r/dtaniwaki/k8s-virtual-device-plugin)

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
