FROM mcr.microsoft.com/playwright:v1.44.0-jammy
WORKDIR /e2e
COPY package*.json ./
RUN npm ci
COPY . .
ENV CI=true
ENTRYPOINT ["npx", "playwright", "test", "--reporter=line", "--project=chromium"]
