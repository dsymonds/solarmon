# Build this with
#   docker build -t dsymonds/solarmon .

FROM golang:1.24-alpine3.21 AS build

WORKDIR /go/src/solarmon
COPY go.mod go.sum ./
RUN go mod download
RUN go install -v \
  github.com/prometheus/client_golang/prometheus

COPY . .
RUN go build -o solarmon -v

# -----

FROM alpine:3.21 AS runtime

# Include tzdata for tracking local times.
RUN apk add --no-cache tzdata

COPY --from=build /go/src/solarmon/solarmon /
ENTRYPOINT ["/solarmon"]
