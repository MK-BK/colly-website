#!/bin/bash
go build -o colly-website main.go

docker build -t colly-website:$1 --no-cache .

docker tag colly-website:$1  wardknight/colly-website:$1

docker push wardknight/colly-website:$1

rm colly-website
