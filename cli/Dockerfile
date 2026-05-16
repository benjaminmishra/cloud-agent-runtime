FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/rune ./cli

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/rune /rune
ENTRYPOINT ["/rune"]
