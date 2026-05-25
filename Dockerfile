FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend ./
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /app/backend
RUN apk add --no-cache ca-certificates
COPY backend/go.mod backend/go.sum* ./
RUN go mod download
COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /pixelmapper .

FROM alpine:3.22
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=backend /pixelmapper /app/pixelmapper
COPY --from=frontend /app/frontend/dist /app/frontend/dist
ENV PORT=8080
ENV DATABASE_PATH=/data/pixelmapper.db
ENV FRONTEND_DIST=/app/frontend/dist
VOLUME ["/data"]
EXPOSE 8080
CMD ["/app/pixelmapper"]
