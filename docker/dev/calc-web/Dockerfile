FROM node:24-alpine3.24

RUN apk add --no-cache make && \
    npm install -g pnpm@12

WORKDIR /workspace/apps/calc-web

COPY apps/calc-web/package.json apps/calc-web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

CMD ["make", "debug"]
