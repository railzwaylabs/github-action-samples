# syntax=docker/dockerfile:1
FROM node:22-alpine AS build
WORKDIR /workspace

RUN npm install --global pnpm@12.3.4
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY apps/sample/package.json apps/sample/package.json
RUN pnpm install --frozen-lockfile --filter github-actions-sample-web...

COPY apps/sample/ apps/sample/
ARG VITE_API_BASE_URL=http://localhost:8080
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
RUN pnpm --filter github-actions-sample-web build

FROM nginx:1.29-alpine
COPY apps/sample/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /workspace/apps/sample/dist /usr/share/nginx/html
EXPOSE 80
