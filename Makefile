# subdirectory in which the actual application resides
APP_DIR := app

# the name of the executable to generate
EXECUTABLE := gofinance

# used for pushing releases
GITHUB_USER := aktau
GITHUB_REPO := gofinance
GITHUB_RELEASE := github-release

# todo: error out of the current git state is not the same as the tag,
# otherwise you're pushing a binary that doesn't belong with the tag...
LAST_TAG := $(shell git describe --abbrev=0 --tags)

UPLOAD_CMD = $(GITHUB_RELEASE) upload -u $(USER) -r $(EXECUTABLE) -t $(LAST_TAG) -n $(subst /,-,$(FILE)) -f bin/$(FILE)

# all executables to build when pushing releases, only include the amd64
# binaries, otherwise the github release will become too big
UNIX_EXECUTABLES := \
	darwin/amd64/$(EXECUTABLE) \
	freebsd/amd64/$(EXECUTABLE) \
	linux/amd64/$(EXECUTABLE)
WIN_EXECUTABLES := \
	windows/amd64/$(EXECUTABLE).exe

COMPRESSED_EXECUTABLES=$(UNIX_EXECUTABLES:%=%.tar.bz2) $(WIN_EXECUTABLES:%.exe=%.zip)
COMPRESSED_EXECUTABLE_TARGETS=$(COMPRESSED_EXECUTABLES:%=bin/%)

# the default make target just builds the executable
all: $(EXECUTABLE)

bin/linux/arm/5/$(EXECUTABLE):
	cd $(APP_DIR) && GOARM=5 GOARCH=arm GOOS=linux go build -o "../$@"
bin/linux/arm/7/$(EXECUTABLE):
	cd $(APP_DIR) && GOARM=7 GOARCH=arm GOOS=linux go build -o "../$@"
bin/linux/386/$(EXECUTABLE):
	cd $(APP_DIR) && GOARCH=386 GOOS=linux go build -o "../$@"
bin/linux/amd64/$(EXECUTABLE):
	cd $(APP_DIR) && GOARCH=amd64 GOOS=linux go build -o "../$@"

bin/darwin/amd64/$(EXECUTABLE):
	cd $(APP_DIR) && GOARCH=amd64 GOOS=darwin go build -o "../$@"

bin/freebsd/386/$(EXECUTABLE):
	cd $(APP_DIR) && GOARCH=386 GOOS=freebsd go build -o "../$@"
bin/freebsd/amd64/$(EXECUTABLE):
	cd $(APP_DIR) && GOARCH=amd64 GOOS=freebsd go build -o "../$@"

bin/windows/386/$(EXECUTABLE):
	cd $(APP_DIR) && GOARCH=386 GOOS=windows go build -o "../$@"
bin/windows/amd64/$(EXECUTABLE):
	cd $(APP_DIR) && GOARCH=amd64 GOOS=windows go build -o "../$@"

# compressed artifacts, makes a huge difference (Go executable is ~9MB,
# after compressing ~2MB)
%.tar.bz2: %
	tar -jcvf "$<.tar.bz2" "$<"
%.zip: %.exe
	zip "$@" "$<"

# git tag -a v$(RELEASE) -m 'release $(RELEASE)'
release: bin/tmp/$(EXECUTABLE) $(COMPRESSED_EXECUTABLE_TARGETS)
	@echo Tagging...
	git push && git push --tags
	@echo Making github release...
	$(GITHUB_RELEASE) release -u $(GITHUB_USER) -r $(GITHUB_REPO) \
		-t $(LAST_TAG) -n $(LAST_TAG) || true
	@echo Uploading...
	$(foreach FILE,$(COMPRESSED_EXECUTABLES),$(UPLOAD_CMD);)
	@echo Done

# update/install all dependencies
dep:
	cd $(APP_DIR) && go list -f '{{join .Deps "\n"}}' | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | xargs go get -u

$(EXECUTABLE): dep
	cd $(APP_DIR) && go build -o "$@"
	@echo $(PWD) $(shell pwd)
	mv "$(APP_DIR)/$@" "./$@"

install:
	cd $(APP_DIR) && go install

clean:
	rm go-app || true
	rm $(EXECUTABLE) || true
	rm -rf bin/

.PHONY: clean release dep install
