FROM ubuntu:focal

COPY colly /usr/bin

EXPOSE 10086

ENTRYPOINT [ "colly" ]