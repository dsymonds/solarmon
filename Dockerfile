# Build this with
#   docker build -t dsymonds/solarmon .

FROM golang:1.22-alpine AS build

WORKDIR /go/src/solarmon
COPY go.mod go.sum ./
RUN go mod download
RUN go build -v \
  github.com/prometheus/client_golang/prometheus

COPY . .
RUN go build -o solarmon -v

# -----

FROM alpine:3.18 AS runtime

COPY --from=build /go/src/solarmon/solarmon /
ENTRYPOINT ["/solarmon"]
