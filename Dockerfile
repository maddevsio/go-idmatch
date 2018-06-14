FROM ubuntu:16.04

LABEL maintainer="rock@maddevs.io"

ARG GO_VERSION="1.10.2"

ENV GOPATH="/go"
ENV PATH=$PATH:/usr/local/go/bin:/go/bin

ENV CGO_CPPFLAGS="-I/usr/local/include"
ENV CGO_LDFLAGS="-L/usr/local/lib -lopencv_core -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_objdetect -lopencv_features2d -lopencv_xfeatures2d"
# ENV CGO_CXXFLAGS="--std=c++1z"

RUN apt-get update && apt-get -y install \
    sudo \
    build-essential \
    wget \
    cmake \
    git \
    unzip \
    libgtk2.0-dev \
    pkg-config \
    libavcodec-dev \
    libavformat-dev \
    libswscale-dev \
    tesseract-ocr-dev \
    tesseract-ocr-eng \
    tesseract-ocr-kir \
    tesseract-ocr-rus \
    libleptonica-dev \
    liblept5 \
    libtbb2 \
    libtbb-dev \
    libjpeg-dev \
    libpng-dev \
    libtiff-dev \
    libjasper-dev \
    libdc1394-22-dev \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

RUN go get -v -u -d gocv.io/x/gocv
# RUN git clone https://gocv.io/x/gocv /go/src/gocv.io/x/gocv
WORKDIR /go/src/gocv.io/x/gocv
RUN make download
RUN make build
RUN make sudo_install
RUN make clean
# RUN make install
# RUN go install gocv.io/x/gocv

WORKDIR /go/src/github.com/maddevsio/go-idmatch
COPY . .
# Temporary workaround
RUN wget https://github.com/tzununbekov/gocv/blob/master/contrib/xfeatures2d.go -P /go/src/gocv.io/x/gocv/contrib

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080

CMD ["go-idmatch", "service"]