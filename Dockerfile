FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /gitbundler .

FROM alpine:3.21
RUN apk add --no-cache git
COPY --from=build /gitbundler /usr/local/bin/gitbundler
ENTRYPOINT ["gitbundler"]
