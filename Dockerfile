# syntax=docker/dockerfile:1

FROM node:20-alpine AS build
WORKDIR /app

# Install deps (uses lockfile for reproducible builds)
COPY package.json package-lock.json ./
RUN npm ci

# Build
COPY . .
RUN npm run build-only

FROM nginx:1.27-alpine

# Serve SPA on a non-conflicting port (80/443 already used on your server)
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/ >/dev/null 2>&1 || exit 1

CMD ["nginx", "-g", "daemon off;"]
