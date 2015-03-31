# Simple dashboard

A very simple dashboard frontend thing in ReactJS and web(sockets) server in Go.  Not yet feature complete and probably not yet usable :)

Currently displays some text with a specified background colour.  This text could be the name of the project and the colour the status the project is in.


## Installing

Build the Javascript file:

```
$ jsx src/jsx docker/html/js
```

Build the server binary:

```
$ go get github.com/go-zoo/bone
$ go get github.com/gorilla/websocket
$ go get gopkg.in/mgo.v2
$ go get gopkg.in/mgo.v2/bson
$ go build -o ../../docker/dashboard src/go/server.go
```

Optional: edit the HTML/CSS in `docker/html`.

Start the application:

```
$ docker-compose up -d
```

Server should now be up and running on port 3000.

### Caveats

No authentication or token access is (_yet_) provided, so you probably shouldn't run this server on a public IP address.

The mongo database files are kept within the container; so if the container is destroyed so are the files.  If you want to keep some history then you need to mount mongo's `/data/db` volume.

## Posting a new dashboard entry

Each status you use should match up to the css class defined in `css.css`.  The current classes are:

* success
* running
* fail

```
$ curl -X POST -d "{\"name\": \"project\", \"status\": \"running\"}" localhost:3000/entry/build
```