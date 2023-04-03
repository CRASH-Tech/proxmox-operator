FROM golang:alpine3.17 as builder

COPY cmd/ /app/cmd
COPY main.go /app/main.go
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum

RUN cd /app; go build -o proxmox-operator

FROM alpine:3.17
COPY --from=builder /app/proxmox-operator /app/proxmox-operator

CMD /app/proxmox-operator
