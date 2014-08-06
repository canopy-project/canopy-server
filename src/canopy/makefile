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
	GOPATH=$$(cd ~/golang; pwd):$$(cd ../../; pwd) go build canopy-cloud-service.go canopy_cloud.go canopy_json.go
	mkdir -p build/
	mv canopy-cloud-service build/

build/canodevtool:
	GOPATH=$$(cd ~/golang; pwd):$$(cd ../../; pwd) go build canopy_cmd.go
	mkdir -p build/
	mv canopy_cmd build/canodevtool

.PHONY: install
install:
	cp $(TARGETS) /usr/local/bin

.PHONY: clean
clean:
	rm -rf build/
