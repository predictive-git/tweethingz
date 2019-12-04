FROM golang:latest as builder

WORKDIR /tmp/
COPY . /tmp/

ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags netgo \
    -ldflags '-w -extldflags "-static"' \
    -mod vendor \
    -o ./service \
    ./src/

FROM gcr.io/distroless/static
COPY --from=builder /src/service .
COPY --from=builder /src/static/ ./static/
COPY --from=builder /src/template/ ./template/

ENTRYPOINT ["/service"]