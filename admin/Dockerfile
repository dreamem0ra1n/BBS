FROM node:22-alpine
WORKDIR /usr/src/app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --ignore-engines

COPY . ./

CMD ["yarn", "serve"]
