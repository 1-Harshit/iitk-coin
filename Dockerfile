# Start with a an GO image.
FROM golang:1.16

LABEL maintainer="Harshit <harshitr20@iitk.ac.in>"

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/1-Harshit/iitk-coin

# Some stuff that everyone has been copy-pasting
# since the dawn of time.

COPY go.mod .

COPY go.sum .

RUN go mod download

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Build the executable
RUN go build

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD [ "./iitk-coin" ]