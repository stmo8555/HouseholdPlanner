# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25 AS build
WORKDIR /app

# Cache mounts persist the module and build caches across builds, so a rebuild
# recompiles only what changed instead of all 189 packages.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o HouseholdPlanner .

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=build /app/HouseholdPlanner .
COPY --from=build /app/web ./web
COPY --from=build /app/food_category_lookup.json .
EXPOSE 8080
CMD ["./HouseholdPlanner"]
