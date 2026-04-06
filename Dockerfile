# Build stage
FROM golang:latest AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o HouseholdPlanner .

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=build /app/HouseholdPlanner .
EXPOSE 8080
CMD ["./HouseholdPlanner"]
