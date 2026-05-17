FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/rune-backend ./cmd/rune-backend

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/rune-backend /rune-backend
EXPOSE 8080
ENTRYPOINT ["/rune-backend"]
