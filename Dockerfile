FROM golang:1.10 as builder

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
RUN go get github.com/olegchorny/chaos-operator/pkg/apis/chaos/v1
RUN go get github.com/olegchorny/chaos-operator/pkg/client/clientset/versioned
##RUN go get github.com/olegchorny/chaos-operator/pkg/client/informers/externalversions/chaos/v1

WORKDIR /go/src/chaos-operator
RUN go build .

##EXPOSE 8080

##FROM alpine:latest  
##RUN apk --no-cache add ca-certificates
##WORKDIR /root/
##COPY --from=builder /go/src/chaos-operator/chaos-operator .

##CMD ["./chaos-operator"]  

FROM ubuntu:16.04
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/cache/apt
# NB: you may need to update RBAC permissions when upgrading kubectl - see kured-rbac.yaml for details
#ADD https://storage.googleapis.com/kubernetes-release/release/v1.10.3/bin/linux/amd64/kubectl /usr/bin/kubectl
#RUN chmod 0755 /usr/bin/kubectl
##WORKDIR /root/
COPY --from=builder /go/src/chaos-operator/chaos-operator /usr/bin/chaos-operator
##COPY ./kured /usr/bin/kured

##CMD ["/go/src/chaos-operator/chaos-operator"]
CMD ["/usr/bin/chaos-operator"]
