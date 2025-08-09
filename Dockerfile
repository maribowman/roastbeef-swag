FROM golang:1.24.6-alpine3.22 AS builder
RUN apk update && apk --no-cache add build-base
LABEL stage=builder
WORKDIR /app
COPY . /app
RUN cd /app
RUN go build -o main .
RUN go test ./... -cover -v

FROM alpine:3.22 as production
RUN apk update && apk --no-cache add ca-certificates
COPY --from=builder /app/main .
COPY /configs /configs/
ENTRYPOINT [ "main" ]

