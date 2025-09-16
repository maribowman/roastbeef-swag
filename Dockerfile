FROM golang:1.25.1-alpine3.22 AS builder
LABEL stage=builder
WORKDIR /src
COPY . .
RUN go test ./... -cover -v
RUN go build -ldflags="-s -w" -o /app/main .

FROM alpine:3.22
RUN apk update && apk --no-cache add ca-certificates
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/main /app/main
# TODO: Might be an issue to only copy the binary -> cannot find go.mod in config
COPY --from=builder --chown=nonroot:nonroot /configs /configs
VOLUME /data
EXPOSE 8800
ENTRYPOINT [ "/app/main" ]
