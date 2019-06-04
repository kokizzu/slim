# `grep -v` does not work on travis. No time to find out why -- xp 2019 Feb 22
PKGS := $(shell go list ./... | grep -v "^github.com/openacid/slim/\(vendor\|prototype\|iohelper\|serialize\|version\)")

# PKGS := github.com/openacid/slim/array \
#         github.com/openacid/slim/bit \
#         github.com/openacid/slim/trie \

SRCDIRS := $(shell go list -f '{{.Dir}}' $(PKGS))

# gofmt check vendor dir. we need to skip vendor manually
GOFILES := $(shell find $(SRCDIRS) -not -path "*/vendor/*" -name "*.go")
GO := go

check: test vet gofmt misspell unconvert staticcheck ineffassign unparam

travis: test vet gofmt misspell unconvert ineffassign unparam coveralls

test:
	$(GO) test $(PKGS)

vet: | test
	$(GO) vet $(PKGS)

staticcheck:
	$(GO) get honnef.co/go/tools/cmd/staticcheck
	# ST1016: methods on the same type should have the same receiver name
	#         .pb.go have this issue.
	staticcheck -checks all,-ST1016 $(PKGS)

misspell:
	$(GO) get github.com/client9/misspell/cmd/misspell
	find $(SRCDIRS) -name '*.go' -or -name '*.md' | grep -v "\bvendor/" | xargs misspell \
		-locale US \
		-error
	misspell \
		-locale US \
		-error \
		*.md *.go

unconvert:
	$(GO) get github.com/mdempsky/unconvert
	unconvert -v $(PKGS)

ineffassign:
	$(GO) get github.com/gordonklaus/ineffassign
	find $(SRCDIRS) -name '*.go' | grep -v "\bvendor/" | xargs ineffassign

pedantic: check errcheck

unparam:
	$(GO) get mvdan.cc/unparam
	unparam ./...

errcheck:
	$(GO) get github.com/kisielk/errcheck
	errcheck $(PKGS)

gofmt:
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(GOFILES) | tee /dev/stderr)"

ben: test
	$(GO) test ./... -run=none -bench=. -benchmem

gen:
	$(GO) generate ./...

readme:
	python ./scripts/build_readme.py
	# brew install nodejs
	# npm install -g doctoc
	doctoc --title '' --github README.md

fix:
	gofmt -s -w $(GOFILES)
	unconvert -v -apply $(PKGS)


coveralls:
	$(GO) get golang.org/x/tools/cmd/cover
	$(GO) get github.com/mattn/goveralls
	$(GO) test -v -covermode=count -coverprofile=coverage.out $(PKGS)
	$$HOME/gopath/bin/goveralls -coverprofile=coverage.out -service=travis-ci -repotoken $$COVERALLS_TOKEN
