FROM golang:1.25.5-alpine3.23 AS builder
LABEL stage=builder
WORKDIR /src
COPY . .
RUN go test ./... -cover -v
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /main .

FROM alpine:3.23
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Europe/Berlin
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot
WORKDIR /
COPY --from=builder --chown=nonroot:nonroot /main /main
COPY --chown=nonroot:nonroot /configs /configs
VOLUME /data
EXPOSE 8800
ENTRYPOINT [ "/main" ]
