ARG GOLANG_IMAGE_TAG="1.20.6-alpine"
FROM golang:${GOLANG_IMAGE_TAG} as base

ARG PDF2IMG_VERSION=dev

# So using 8.13.3 for now

ARG VIPS_VERSION="8.13.3"
RUN apk update && apk add --no-cache \
    automake build-base pkgconfig glib-dev gobject-introspection \
    libxml2-dev expat-dev jpeg-dev libwebp-dev libpng-dev \
    libjpeg-turbo-dev 
ARG VIPS_URL=https://github.com/libvips/libvips/releases/download

ADD ${VIPS_URL}/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz \
    /tmp/

RUN cd /tmp \
    && tar xf vips-${VIPS_VERSION}.tar.gz \
    && cd vips-${VIPS_VERSION} \
    && CFLAGS="-g -O3" CXXFLAGS="-D_GLIBCXX_USE_CXX11_ABI=0 -g -O3" \
    ./configure \
        --disable-debug \
        --disable-dependency-tracking \
        --disable-introspection \
        --disable-static \
        --without-gsf \
        --without-magick \
        --without-openslide \
        --without-pdfium \
        --enable-gtk-doc-html=no \
        --enable-gtk-doc=no \
        --enable-pyvips8=no \
        && make -j4 V=0 \
        && make install \
        && ldconfig ; exit 0


# create a working directory inside the image
WORKDIR /go/src/github.com/felixgao/pdf2img
# Cache go modules
ENV GO111MODULE="on"
ENV CPATH="/usr/local/include"
ENV LIBRARY_PATH="/usr/local/lib"
ENV PKG_CONFIG_PATH="/vips/lib/pkgconfig:/usr/local/lib/pkgconfig:/usr/lib/pkgconfig:$PKG_CONFIG_PATH"
ENV GOROOT="/usr/local/go"
ENV GOPATH="/go"
ENV PATH="/go/bin:$PATH"

# Copy go mod and sum files and files i.e all files ending with .go 
COPY go.mod go.sum *.go  ./


# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

RUN go list -mod=readonly -u -m -json all 
# Build the Go app with ldflags
RUN CGO_ENABLED=1 GOOS=linux go build -a \
# flag -s disable symbol table, -w disable DWARF generation, -h halt on error
    -ldflags="-s -w -h -X main.Version=${PDF2IMG_VERSION}" \
    -o pdf2png && apk del automake build-base


# build target image
FROM golang:${GOLANG_IMAGE_TAG} as prod
RUN apk --update add --no-cache \
    fftw glib expat libjpeg-turbo libpng \
	libwebp giflib librsvg libgsf libexif lcms2
WORKDIR /root/
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# This location contains the libvips that is built in the base
COPY --from=base /usr/local/lib/* /usr/local/lib/

WORKDIR /app/
COPY --from=base /go/src/github.com/felixgao/pdf2img/pdf2png .

# add a section on the user that is non-root to run this app

#tell docker what port to expose
EXPOSE 8080

# Run the binary program produced by `go install`
CMD ["./pdf2png"]

