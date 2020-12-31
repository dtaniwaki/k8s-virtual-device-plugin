# Builder phase.
FROM golang:1.15.6 AS builder

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install k8s-virtual-device-plugin

FROM alpine:3.12.3

COPY device.yaml /etc/k8s-virtual-device-plugin/device.yaml
COPY --from=build /go/bin/k8s-virtual-device-plugin /bin/k8s-virtual-device-plugin
CMD ["/bin/k8s-virtual-device-plugin", "/etc/k8s-virtual-device-plugin/device.yaml"]
