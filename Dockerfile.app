# Application image. The dev-container image lives in ./Dockerfile; this one
# builds the deployable server. Templates and static assets are embedded via
# embed.FS, so the final image needs nothing but the binary.
FROM golang:1.26.3-alpine AS build
WORKDIR /src

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /app/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot AS run
COPY --from=build /app/server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
