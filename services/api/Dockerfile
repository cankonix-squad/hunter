FROM golang:1.26.9-alpine AS build
ARG TARGET=api
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/${TARGET}

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app
COPY --from=build /out/service /app/service
USER app
EXPOSE 8080
ENTRYPOINT ["/app/service"]
