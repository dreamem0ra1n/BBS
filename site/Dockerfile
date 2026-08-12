FROM node:22-alpine

# env
ENV HOST=0.0.0.0
ENV APP_HOME=/code/bbs-go/site
ENV NODE_ENV=production
ENV NODE_OPTIONS=--max-old-space-size=8192

# code
WORKDIR "$APP_HOME"
COPY package.json yarn.lock ./

# install
RUN yarn install --frozen-lockfile --ignore-engines --production=false

COPY . ./
RUN yarn build
# ⬆️ build

EXPOSE 3000
CMD ["yarn", "start"]
