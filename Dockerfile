FROM golang:alpine3.17 as builder

COPY cmd/ /app/cmd
COPY main.go /app/main.go
COPY utils.go /app/utils.go
COPY leaderElection.go /app/leaderElection.go
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
WORKDIR /app
RUN go build -o proxmox-operator

FROM alpine:3.17
COPY --from=builder /app/proxmox-operator /app/proxmox-operator
WORKDIR /app
CMD /app/proxmox-operator
