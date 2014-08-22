GOPATH := $(shell cd ../../..; pwd)
export GOPATH

PROJECT = launchpad.net/account-polld

ifneq ($(CURDIR),$(GOPATH)/src/launchpad.net/account-polld)
$(error unexpected curdir and/or layout)
endif

GODEPS = launchpad.net/gocheck
GODEPS += launchpad.net/go-dbus/v1
GODEPS += launchpad.net/go-xdg/v0
GODEPS += launchpad.net/ubuntu-push

TOTEST = $(shell env GOPATH=$(GOPATH) go list $(PROJECT)/...)

check:
	go test $(TESTFLAGS) $(TOTEST)

check-format:
	scripts/check_fmt $(PROJECT)

build:
	go build launchpad.net/account-polld/cmd/account-polld

format:
	go fmt $(PROJECT)/...

# very basic cleanup stuff; needs more work
clean:
	rm account-polld
 
bootstrap:
	rm -r $(GOPATH)/pkg
	mkdir -p $(GOPATH)/bin
	mkdir -p $(GOPATH)/pkg
	go get -d -u $(GODEPS)
	go install $(GODEPS)

.PHONY: bootstrap check check-format format build clean

