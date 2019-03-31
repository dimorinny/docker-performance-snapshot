# Docker performance snapshot

Tool for getting resource usage statistics from docker container. Currently supported statistics:
* CPU usage (in percentage)
* RAM usage (in percentage)

## Usage

#### Installing

```bash
go get github.com/dimorinny/docker-performance-snapshot
go install github.com/dimorinny/docker-performance-snapshot
```

Also remember to actually put $GOPATH/bin into your regular $PATH

#### Starting

```bash
CONTAINER="<running-container-id>" RESULT_DIRECTORY="<directory-for-results>" docker-performance-snapshot
```

Or you can use docker container:

```bash
mkdir /reports

docker run -d \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /reports:/reports \
    -e RESULT_DIRECTORY="/reports" \
    -e CONTAINER="<running-container-id>" \
    dimorinny/docker-performance-snapshot:latest
```