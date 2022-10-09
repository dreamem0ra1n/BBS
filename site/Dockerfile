# FROM node:16.15.1-alpine
# WORKDIR /usr/src/app
# COPY . ./
# RUN npm install -g --registry=https://r.npm.taobao.org

# CMD ["npm", "run", "dev"]

FROM node:16.15.1-alpine

# env
ENV HOST 0.0.0.0
ENV APP_HOME /code/bbs-go/site
ENV NODE_ENV production

# code
WORKDIR "$APP_HOME"
COPY . ./
COPY nuxt.config.js ./nuxt.config.js

# install
RUN npm config set sass_binary_site=https://npm.taobao.org/mirrors/node-sass \
  && npm install -g --registry=https://r.npm.taobao.org \
  && npm run build
# ⬆️ build

EXPOSE 3000
CMD ["npm", "start"]