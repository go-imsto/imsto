.SILENT :
.PHONY : dep vet main clean dist package
DATE := `date '+%Y%m%d'`

WITH_ENV = env `cat .env 2>/dev/null | xargs`

ORIG:=liut7
NAME:=imsto
ROOF:=github.com/go-imsto/$(NAME)
SOURCES=$(shell find cmd config image rpc storage -type f \( -name "*.go" ! -name "*_test.go" \) -print )
STATICS=$(shell find apps/static -type f \( -name "*.css" -o -name "*.js" -o -name "*.png" -o -name "*.gif" \) -print )
TAG:=`git describe --tags --always`
LDFLAGS:=-X $(ROOF)/config.Version=$(TAG)-$(DATE)
GO=$(shell which go)
VET=$(GO) vet -all
GOMOD=$(shell echo "$${GO111MODULE:-auto}")


main:
	echo "Building $(NAME)"
	CGO_ENABLED=1 GO111MODULE=$(GOMOD) $(GO) build -v -ldflags "$(LDFLAGS)" .

dep:
	GO111MODULE=$(GOMOD) $(GO) install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	GO111MODULE=$(GOMOD) $(GO) get github.com/liut/staticfiles

vet:
	echo "Checking ./..."
	$(WITH_ENV) GO111MODULE=$(GOMOD) $(VET) ./cmd/... ./config ./image/... ./rpc/... ./storage/... ./web/...

clean:
	echo "Cleaning dist"
	rm -rf dist
	rm -f $(NAME) $(NAME)-*

dist/linux_amd64/$(NAME): $(SOURCES)
	echo "Building $(NAME) of linux"
	mkdir -p dist/linux_amd64 && cd dist/linux_amd64 && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO111MODULE=$(GOMOD) $(GO) build -ldflags "$(LDFLAGS) -s -w" $(ROOF)
	ls -l $@

dist/darwin_amd64/$(NAME): $(SOURCES)
	echo "Building $(NAME) of darwin"
	mkdir -p dist/darwin_amd64 && cd dist/darwin_amd64 && CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 GO111MODULE=$(GOMOD) $(GO) build -ldflags "$(LDFLAGS) -w" $(ROOF)
	ls -l $@

dist: vet dist/linux_amd64/$(NAME) dist/darwin_amd64/$(NAME)

package-linux: dist/linux_amd64/$(NAME)
	tar -cvJf $(NAME)-linux-amd64-$(TAG).tar.xz -C dist/linux_amd64 $(NAME)

package-macos: dist/darwin_amd64/$(NAME)
	tar -cvJf $(NAME)-darwin-amd64-$(TAG).tar.xz -C dist/darwin_amd64 $(NAME)

package: package-linux package-macos
	ls -l $(NAME)-*.tar.?z

admin:
	echo "Building $@"
	mkdir -p dist/linux_amd64 && GOOS=linux GOARCH=amd64 GO111MODULE=$(GOMOD) $(GO) build -ldflags "$(LDFLAGS)" -o dist/linux_amd64/$(NAME)-admin $(ROOF)/apps/$(NAME)-admin
	mkdir -p dist/darwin_amd64 && GOOS=darwin GOARCH=amd64 GO111MODULE=$(GOMOD) $(GO) build -ldflags "$(LDFLAGS)" -o dist/darwin_amd64/$(NAME)-admin $(ROOF)/apps/$(NAME)-admin

test-image:
	$(VET) ./image
	mkdir -p tests
	$(WITH_ENV) $(GO) test -v -cover -coverprofile tests/cover_image.out ./image
	$(WITH_ENV) $(GO) tool cover -html=tests/cover_image.out -o tests/cover_image.out.html

test-storage:
	$(VET) ./storage
	mkdir -p tests
	$(WITH_ENV) $(GO) test -v -cover -coverprofile tests/cover_storage.out ./storage
	$(WITH_ENV) $(GO) tool cover -html=tests/cover_storage.out -o tests/cover_storage.out.html

test-rpc:
	$(VET) ./rpc
	mkdir -p tests
	$(WITH_ENV) $(GO) test -v -cover -coverprofile tests/cover_rpc.out ./rpc
	$(WITH_ENV) $(GO) tool cover -html=tests/cover_rpc.out -o tests/cover_rpc.out.html

generate:
	GO111MODULE=$(GOMOD) $(GO) generate ./...

static: $(STATICS)
	echo 'packing UI files into static'
	staticfiles --package static -o web/admin/static/files.go ./apps/static
.PHONY: $@


docker-db-build:
	echo "Building database image"
	docker build -t $(ORIG)/$(NAME)-db:$(TAG) database/
	docker tag $(ORIG)/$(NAME)-db:$(TAG) $(ORIG)/$(NAME)-db:latest
	docker save -o $(ORIG)_$(NAME)_db.tar $(ORIG)/$(NAME)-db:$(TAG) $(ORIG)/$(NAME)-db:latest && gzip -9f $(ORIG)_$(NAME)_db.tar
.PHONY: $@

docker-build:
	echo "Building docker image"
	docker build --rm -t $(ORIG)/$(NAME):$(TAG) .
	docker tag $(ORIG)/$(NAME):$(TAG) $(ORIG)/$(NAME):latest
	docker save -o $(ORIG)_$(NAME).tar $(ORIG)/$(NAME):$(TAG) $(ORIG)/$(NAME)-db:latest && gzip -9f $(ORIG)_$(NAME).tar
.PHONY: $@
