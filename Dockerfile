FROM golang:1.26.3-alpine3.23 AS builder
LABEL stage=builder
WORKDIR /src
RUN apk add --update gcc musl-dev
COPY . .
RUN go test ./... -cover -v
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /main .

FROM alpine:3.23
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Europe/Berlin
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
WORKDIR /
COPY --from=builder --chown=nonroot:nonroot /main /main
COPY /configs /configs
RUN mkdir /data && chown nonroot:nonroot /data
USER nonroot
VOLUME /data
EXPOSE 8800
ENTRYPOINT [ "/main" ]
