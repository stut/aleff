FROM golang:1.23.1-alpine3.20 AS build

WORKDIR /app

RUN adduser -D scratchuser

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY src/*.go ./

RUN CGO_ENABLED=0 go build -o /aleff

FROM scratch

WORKDIR /

USER scratchuser

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /aleff /aleff

ENTRYPOINT ["/aleff"]

