.SILENT :
.PHONY : dep vet main clean dist package
DATE := `date '+%Y%m%d'`

WITH_ENV = env `cat .env 2>/dev/null | xargs`

ORIG:=liut7
NAME:=imsto
ROOF:=github.com/go-imsto/$(NAME)
SOURCES=$(shell find base cmd config image storage -type f \( -name "*.go" ! -name "*_test.go" \) -print )
TAG:=`git describe --tags --always`
LDFLAGS:=-X $(ROOF)/cmd.VERSION=$(TAG)-$(DATE)

main:
	echo "Building $(NAME)"
	go build -ldflags "$(LDFLAGS)" .

dep:
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow

vet:
	echo "Checking ./..."
	go vet -vettool=$(which shadow) -atomic -bool -copylocks -nilfunc -printf -rangeloops -unreachable -unsafeptr -unusedresult ./...

clean:
	echo "Cleaning dist"
	rm -rf dist
	rm -f $(NAME) $(NAME)-*

dist/linux_amd64/$(NAME): $(SOURCES)
	echo "Building $(NAME) of linux"
	mkdir -p dist/linux_amd64 && cd dist/linux_amd64 && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s -w" $(ROOF)

dist/darwin_amd64/$(NAME): $(SOURCES)
	echo "Building $(NAME) of darwin"
	mkdir -p dist/darwin_amd64 && cd dist/darwin_amd64 && GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS) -w" $(ROOF)

dist: vet dist/linux_amd64/$(NAME) dist/darwin_amd64/$(NAME)

package: dist
	tar -cvJf $(NAME)-linux-amd64-$(TAG).tar.xz -C dist/linux_amd64 $(NAME)
	tar -cvJf $(NAME)-darwin-amd64-$(TAG).tar.xz -C dist/darwin_amd64 $(NAME)


docker-db-build:
	echo "Building database image"
	docker build -t $(ORIG)/$(NAME)-db:$(TAG) database/
	docker tag $(ORIG)/$(NAME)-db:$(TAG) $(ORIG)/$(NAME)-db:latest
	docker save -o $(ORIG)_$(NAME)_db.tar $(ORIG)/$(NAME)-db:$(TAG) $(ORIG)/$(NAME)-db:latest && gzip -9f $(ORIG)_$(NAME)_db.tar
.PHONY: $@

docker-build:
	echo "Building docker image"
	docker build -t $(ORIG)/$(NAME):$(TAG) .
	docker tag $(ORIG)/$(NAME):$(TAG) $(ORIG)/$(NAME):latest
.PHONY: $@