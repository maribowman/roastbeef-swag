FROM golang:1.21.1-alpine3.18 AS builder
RUN apk update && apk --no-cache add build-base
LABEL stage=builder
WORKDIR /building-site
COPY . /building-site
RUN cd /building-site
RUN go build -o main .
RUN go test ./... -cover -v

FROM alpine:3.18 as production
RUN apk update && apk --no-cache add ca-certificates
COPY --from=builder /building-site/main .
COPY /configs /configs/
ENTRYPOINT ./main
