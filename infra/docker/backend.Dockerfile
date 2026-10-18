FROM golang:1.27.2-alpine AS build

ARG TARGET=api

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/nuvro ./cmd/${TARGET}

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=build /out/nuvro /nuvro

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/nuvro"]
