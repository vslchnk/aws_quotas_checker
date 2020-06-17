FROM golang:1.13
WORKDIR /go/src/github.com/vslchnk/aws_quotas_checker/
RUN go get -d -v github.com/aws/aws-sdk-go/...
COPY . .
WORKDIR example
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o aws_quota_checker .

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /go/src/github.com/vslchnk/aws_quotas_checker/example/aws_quota_checker .
CMD ["./aws_quota_checker"]
