FROM node:17.9 as build-ui

WORKDIR /work

COPY ui/package.json ./
COPY ui/package-lock.json ./
RUN npm clean-install
COPY ui/ ./
RUN npm run build

FROM golang:1.17-alpine as build-scandoc

WORKDIR /build

COPY go.* ./
RUN go mod download \
    && mkdir -p ui/build

COPY *.go ./
COPY --from=build-ui /work/build ./ui/build

RUN CGO_ENABLED=0 GOOS=linux go build -o scandoc

FROM debian:bullseye-slim as build-jbig2enc

RUN apt-get update && apt-get install -y --no-install-recommends \
    automake \
    build-essential \
    ca-certificates \
    git \
    libleptonica-dev \
    libtool \
    zlib1g-dev

WORKDIR /build

ENV JBIG2ENC_REV ea6a40a

RUN git clone https://github.com/agl/jbig2enc.git \
    && cd jbig2enc \
    && git checkout ${JBIG2ENC_REV} \
    && ./autogen.sh \
    && ./configure \
    && make \
    && mkdir /build/jbig2enc-install \
    && make install DESTDIR=/build/jbig2enc-install

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    img2pdf \
    python3-dev \
    python3-pip \
    python3-pkgconfig \
    sane \
    sane-utils \
    # ocrmypdf
    ghostscript \
    icc-profiles-free \
    libcairo2-dev \
    libxml2 \
    pngquant \
    tesseract-ocr \
    unpaper \
    zlib1g \
    # debug tools
    less \
    procps \
    usbutils \
    vim \
    && rm -rf /var/lib/apt/lists/*

# todo: scanner specific setup... separate this out somehow?
ADD https://www.josharcher.uk/static/files/2016/10/1300_0C26.nal /usr/share/sane/epjitsu/1300_0C26.nal
RUN echo epjitsu > /etc/sane.d/dll.conf

ENV OCRMYPDF_VERSION 13.4.3
ENV NOTESHRINK_VERSION 0.1.1

RUN pip install noteshrink==${NOTESHRINK_VERSION} ocrmypdf==${OCRMYPDF_VERSION}

COPY --from=build-jbig2enc /build/jbig2enc-install/ /
COPY --from=build-scandoc /build/scandoc /usr/local/bin/scandoc

EXPOSE 8090

WORKDIR /work

CMD /bin/bash
