FROM golang:1.17-alpine3.14 AS builder

RUN apk update && apk upgrade && \
   apk add --no-cache bash git

ENV GO111MODULE=on

WORKDIR /go/src/service

# docker cache go downloads
COPY go.* ./
RUN go mod download

COPY . ./

#disable crosscompiling
ENV CGO_ENABLED=0

#compile linux only
ENV GOOS=linux

#build the binary with debug information removed
RUN go build -ldflags '-w -s' -a -installsuffix cgo -o /viewer-config-service

FROM scratch as service
EXPOSE 80
WORKDIR /
ENV PATH=/

COPY --from=builder /viewer-config-service /
COPY ./swaggerui/ /swaggerui/

ENTRYPOINT "viewer-config-service"
