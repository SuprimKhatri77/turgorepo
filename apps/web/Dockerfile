# Production image. Build from the repo root, because the web app uses workspace
# packages. Next inlines NEXT_PUBLIC_* at build time, so it's a build argument:
#   docker build -f apps/web/Dockerfile --build-arg NEXT_PUBLIC_API_URL=https://api.example.com -t turgorepo-web .
FROM node:26-alpine AS base

RUN npm install --global bun@1.4.4

FROM base AS deps

WORKDIR /app

COPY bun.lock package.json turbo.json ./
COPY apps/api/package.json ./apps/api/
COPY apps/web/package.json ./apps/web/
COPY packages/api-client/package.json ./packages/api-client/
COPY packages/openapi/package.json ./packages/openapi/
COPY packages/types/package.json ./packages/types/
COPY packages/typescript-config/package.json ./packages/typescript-config/
RUN bun install --frozen-lockfile --ignore-scripts

FROM deps AS build

ARG NEXT_PUBLIC_API_URL
RUN test -n "$NEXT_PUBLIC_API_URL" || { echo "Set the NEXT_PUBLIC_API_URL build argument"; exit 1; }

COPY apps/web ./apps/web
COPY packages ./packages

ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_OUTPUT=standalone
ENV NODE_ENV=production

WORKDIR /app/apps/web
RUN node node_modules/next/dist/bin/next build

FROM node:26-alpine AS run

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

COPY --from=build --chown=node:node /app/apps/web/.next/standalone ./
COPY --from=build --chown=node:node /app/apps/web/.next/static ./apps/web/.next/static
COPY --from=build --chown=node:node /app/apps/web/public ./apps/web/public

USER node

EXPOSE 3000

CMD ["node", "apps/web/server.js"]
