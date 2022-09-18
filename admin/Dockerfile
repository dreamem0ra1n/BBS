FROM node:18-alpine3.15
WORKDIR /usr/src/app
COPY . ./
RUN npm install -g --registry=https://r.npm.taobao.org 

CMD ["npm run serve"]