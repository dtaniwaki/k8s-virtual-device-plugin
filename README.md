# k8s-virtual-device-plugin

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
