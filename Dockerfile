FROM golang:latest AS build-env
RUN mkdir -p /go/webhook
WORKDIR /go/webhook
COPY  . .
RUN useradd -u 10001 webhook
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o webhook

FROM scratch
COPY --from=build-env /go/webhook .
COPY --from=build-env /etc/passwd /etc/passwd
USER webhook
ENTRYPOINT ["/webhook"]
