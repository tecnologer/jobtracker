FROM node:20-alpine AS frontend
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:latest AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o jobtracker .

FROM alpine:latest
WORKDIR /app
COPY --from=backend /app/jobtracker .
COPY --from=frontend /web/dist ./web/dist
EXPOSE 8080
CMD ["./jobtracker"]
