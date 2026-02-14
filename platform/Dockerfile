# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY platform/go.mod platform/go.sum ./
RUN go mod download
COPY platform/ .
RUN CGO_ENABLED=0 go build -o /app/server .

# Run stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
# Mirror the repo layout so ../AMS_sc_terminal_daily_2023.csv resolves from the working dir
COPY --from=build /app/server /app/platform/server
COPY AMS_sc_terminal_daily_2023.csv AMS_sc_terminal_daily_2024.csv /app/
WORKDIR /app/platform
EXPOSE 8000
CMD ["./server"]
