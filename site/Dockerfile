FROM node:16.15.1-alpine
WORKDIR /usr/src/app
COPY . ./
RUN npm install -g --registry=https://r.npm.taobao.org \
npm run build

CMD ["npm run start"]