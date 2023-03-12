FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.18.0-alpine3.15 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN adduser -D scratchuser

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY src/*.go ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /aleff

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch

WORKDIR /

USER scratchuser

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /aleff /aleff

ENTRYPOINT ["/aleff"]

