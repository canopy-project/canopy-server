#
# Copyright 2014 Gregory Prisament
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

TARGETS := build/canopy-cloud-service build/canodevtool

.PHONY: all
.DEFAULT: all
all: $(TARGETS)

.PHONY: $(TARGETS)

build/canopy-cloud-service:
	mkdir -p ~/.canopy/golang
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go build canopy-cloud-service.go canopy_cloud.go canopy_json.go
	mkdir -p build/
	mv canopy-cloud-service build/

build/canodevtool:
	mkdir -p ~/.canopy/golang
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go build canopy_cmd.go
	mkdir -p build/
	mv canopy_cmd build/canodevtool

.PHONY: go_get_deps
go_get_deps:
	mkdir -p ~/.canopy/golang
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go get code.google.com/p/go.net/websocket
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go get github.com/gocql/gocql
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go get github.com/gorilla/sessions
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go get github.com/gorilla/context
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go get github.com/gorilla/mux
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go get github.com/sendgrid/sendgrid-go
	GOPATH=$$(cd ~/.canopy/golang; pwd):$$(cd ../../; pwd) go get code.google.com/p/go.crypto/bcrypt

.PHONY: install
install:
	cp $(TARGETS) /usr/local/bin
	cp ../../init_script/canopy-cloud-service /etc/init.d/
	../../init_script/create-canopy-group-user.sh

.PHONY: clean
clean:
	rm -rf build/
