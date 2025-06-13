APP_NAME := task-compose
GO_MODULE_PATH := $(shell go list -m)

CLI_LDFLAGS := -X '${GO_MODULE_PATH}/app.Version=$$(git describe --tags)' \
           -X '${GO_MODULE_PATH}/app.ExecutionMode=CLI' \
           -X '${GO_MODULE_PATH}/app.BuildDate=$$(date +%Y-%m-%d)' \
           -X '${GO_MODULE_PATH}/app.CommitHash=$$(git rev-parse --short HEAD)'

GUI_LDFLAGS := -X '${GO_MODULE_PATH}/app.Version=$$(git describe --tags)' \
           -X '${GO_MODULE_PATH}/app.ExecutionMode=GUI' \
           -X '${GO_MODULE_PATH}/app.BuildDate=$$(date +%Y-%m-%d)' \
           -X '${GO_MODULE_PATH}/app.CommitHash=$$(git rev-parse --short HEAD)'

build:
	@echo "Building $(APP_NAME) with LDFLAGS: $(LDFLAGS)"
	go build -ldflags="$(CLI_LDFLAGS)" -o $(APP_NAME)
	go build -ldflags="$(GUI_LDFLAGS)" -o $(APP_NAME)-portable

release:
	goreleaser release --clean

prerelease:
	goreleaser release --skip=publish --clean --skip=validate

clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)-portable
	rm -f $(APP_NAME)