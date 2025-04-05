FROM golang:1.24

LABEL org.opencontainers.image.source=https://github.com/led0nk/ark-overseer

COPY . /go/src/github.com/led0nk/ark-overseer

WORKDIR /go/src/github.com/led0nk/ark-overseer

RUN CGO_ENABLED=0 go build -v -o /ark-overseer cmd/api/main.go

FROM scratch

COPY --from=0 /ark-overseer /ark-overseer

EXPOSE 8080

CMD ["/ark-overseer", "-addr", "0.0.0.0:8080"]
