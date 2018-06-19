FROM denismakogon/gocv-alpine:3.4.1-buildstage as build-stage

RUN apk add --update leptonica tesseract-ocr-dev tesseract-ocr-data-rus
RUN go get -u -d github.com/maddevsio/go-idmatch
ADD https://raw.githubusercontent.com/tzununbekov/gocv/master/contrib/xfeatures2d.go $GOPATH/src/gocv.io/x/gocv/contrib
RUN cd $GOPATH/src/github.com/maddevsio/go-idmatch && go build main.go

FROM denismakogon/gocv-alpine:3.4.1-runtime

COPY --from=build-stage /usr/share/tessdata/ /usr/share/tessdata/
COPY --from=build-stage /usr/lib/libgif.so.7 /usr/lib/liblept.so.5 /usr/lib/libtesseract.so.3 /usr/lib/
COPY --from=build-stage /go/src/github.com/maddevsio/go-idmatch /go-idmatch
WORKDIR /go-idmatch

EXPOSE 8080
ENTRYPOINT ["./main", "service"]

