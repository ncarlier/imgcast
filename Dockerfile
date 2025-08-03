#########################################
# Build backend stage
#########################################
FROM golang:1.24 AS backend-builder

# Repository location
ARG REPOSITORY=github.com/ncarlier

# Artifact name
ARG ARTIFACT=imgcast

# Copy sources into the container
ADD . /go/src/$REPOSITORY/$ARTIFACT

# Set working directory
WORKDIR /go/src/$REPOSITORY/$ARTIFACT

# Build the binary
RUN make

#########################################
# Distribution stage
#########################################
FROM gcr.io/distroless/base-debian12:nonroot

# Repository location
ARG REPOSITORY=github.com/ncarlier

# Artifact name
ARG ARTIFACT=imgcast

# Install backend binary
COPY --from=backend-builder /go/src/$REPOSITORY/$ARTIFACT/release/$ARTIFACT /usr/local/bin/$ARTIFACT

# Exposed ports
EXPOSE 8080

# Define entrypoint
ENTRYPOINT [ "imgcast" ]
