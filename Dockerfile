FROM golang:1.12 as builder

WORKDIR /build
COPY . ./

RUN go install ./cmd/control-plane
RUN go install ./cmd/server

WORKDIR /
RUN rm -rf /build

# xds server
EXPOSE 5678

# rest api
EXPOSE 8000

# simple http server
EXPOSE 8080
CMD ["/go/bin/control-plane"]
