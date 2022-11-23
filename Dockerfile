FROM ubuntu:focal

COPY colly-website /usr/bin

EXPOSE 10086

ENTRYPOINT [ "colly-website" ]