FROM golang:1.22

LABEL org.opencontainers.image.source=https://github.com/led0nk/ark-clusterinfo

COPY . /go/src/github.com/led0nk/ark-clusterinfo

WORKDIR /go/src/github.com/led0nk/ark-clusterinfo

RUN CGO_ENABLED=0 go build -v -o /ark-clusterinfo cmd/server/main.go

FROM scratch

COPY --from=0 /ark-clusterinfo /ark-clusterinfo

EXPOSE 8080

CMD ["/ark-clusterinfo", "-addr", "0.0.0.0:8080"]
