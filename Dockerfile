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
# Version resolution: git describe when .git is in the build context; else the
# VERSION build arg (Railway exposes a VERSION service variable as a build arg
# because it's declared here); else "dev".
ARG VERSION=dev
RUN VERSION=$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || echo "$VERSION") \
 && CGO_ENABLED=0 go build -ldflags "-X github.com/tecnologer/jobtracker/handler.buildVersion=$VERSION" -o jobtracker .

FROM alpine:latest
WORKDIR /app
COPY --from=backend /app/jobtracker .
COPY --from=frontend /web/dist ./web/dist
EXPOSE 8080
CMD ["./jobtracker"]
