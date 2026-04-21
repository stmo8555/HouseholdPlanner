# Build stage
FROM golang:latest AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o HouseholdPlanner .

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=build /app/HouseholdPlanner .
COPY --from=build /app/web ./web
EXPOSE 8080
CMD ["./HouseholdPlanner"]
