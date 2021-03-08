#!/bin/bash
go get github.com/hairyhenderson/gomplate

gomplate -f $GOPATH/src/github.com/wetix/centrifugo/misc/proto/api.template.proto > $GOPATH/src/github.com/wetix/centrifugo/misc/proto/api.proto
GOGO=1 gomplate -f $GOPATH/src/github.com/wetix/centrifugo/misc/proto/api.template.proto > $GOPATH/src/github.com/wetix/centrifugo/internal/api/api.proto

cd $GOPATH/src/github.com/wetix/centrifugo/internal/api && protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/wetix/centrifugo/vendor:. --gogofaster_out=plugins=grpc:. api.proto
