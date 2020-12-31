###########
# Builder
###########
FROM golang:1.15.6 AS builder

# ENV GO111MODULE=on
ENV GOPATH=
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

###########
# Prod
###########
FROM alpine:3.12.3

COPY device.yaml /etc/k8s-virtual-device-plugin/device.yaml
COPY --from=builder /go/dist/k8s-virtual-device-plugin /bin/k8s-virtual-device-plugin
CMD ["/bin/k8s-virtual-device-plugin", "/etc/k8s-virtual-device-plugin/device.yaml"]
