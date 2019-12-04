FROM golang:latest as builder

WORKDIR /build/
COPY . /build/

ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags netgo \
    -ldflags "-w -extldflags -static" \
    -mod vendor \
    -o ./service \
    ./src/


FROM gcr.io/distroless/static:nonroot

COPY --from=builder /build/service .
COPY --from=builder /build/static/ ./static/
COPY --from=builder /build/template/ ./template/

ENTRYPOINT ["./service"]
