FROM golang:1.10


RUN mkdir -p /go/src/chaos-operator
ADD . /go/src/chaos-operator
WORKDIR /go/src

RUN go get k8s.io/apimachinery/pkg/api/errors 
RUN go get k8s.io/apimachinery/pkg/apis/meta/v1  
RUN go get k8s.io/client-go/kubernetes 
RUN go get k8s.io/api/core/v1
RUN go get k8s.io/client-go/rest 
RUN go get github.com/Sirupsen/logrus 
##RUN go get -u k8s.io/kubernetes/pkg/util/taints
##RUN go get github.com/robfig/cron
RUN go get gopkg.in/robfig/cron.v2
##RUN go get k8s.io/kubernetes/vendor/k8s.io/api/core/v1

WORKDIR /go/src/chaos-operator
RUN go build .

##EXPOSE 8080

##FROM alpine:latest  
##RUN apk --no-cache add ca-certificates
##WORKDIR /root/
##COPY --from=builder /go/src/chaos-operator/chaos-operator .

##CMD ["./chaos-operator"]  

CMD ["/go/src/chaos-operator/chaos-operator"]
