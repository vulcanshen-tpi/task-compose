# Dockerfile
# --- 階段 1: 下載器 (Downloader Stage) ---
# 這一階段用於從 GitHub 下載 Go 原始碼，避免將 .git 目錄複製到建置環境中
# 使用一個輕量級的基礎影像，包含 git
FROM alpine/git:2.44.0 AS downloader

# 定義一個 ARG，用於接收 goreleaser 傳入的版本號
ARG VERSION

# 設置工作目錄
WORKDIR /tmp

# 克隆儲存庫並切換到指定版本
# git clone https://github.com/vulcanshen/task-compose.git .
# git checkout v${VERSION}
# 為了避免在 CI 環境中遇到認證問題，更推薦直接從本地複製
COPY go.mod go.sum ./
COPY main.go ./
# 複製你的其他 Go 原始碼文件/資料夾
COPY app/ app/
COPY cmd/ cmd/
COPY config/ config/
COPY procedure/ procedure/
COPY utils/ utils/


# --- 階段 2: 建置器 (Builder Stage) ---
# 使用一個包含 Go 編譯環境的基礎影像
FROM golang:1.22-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 從 downloader 階段複製原始碼 (只複製需要的 Go 檔案)
COPY --from=downloader /tmp/go.mod /app/go.mod
COPY --from=downloader /tmp/go.sum /app/go.sum
COPY --from=downloader /tmp/main.go /app/main.go
COPY --from=downloader /tmp/app/ /app/app/
COPY --from=downloader /tmp/cmd/ /app/cmd/
COPY --from=downloader /tmp/config/ /app/config/
COPY --from=downloader /tmp/procedure/ /app/procedure/
COPY --from=downloader /tmp/utils/ /app/utils/

# 修正 go.mod 檔案的權限 (有時在 alpine 下需要)
RUN chmod 644 go.mod && chmod 644 go.sum

# 下載 Go modules 依賴並編譯應用程式
# 注意：這裡的編譯指令需要和 .goreleaser.yaml 中的 builds.ldflags 保持一致
ARG VERSION # 再次聲明 ARG VERSION，供 ldflags 使用
RUN CGO_ENABLED=0 go build -o /usr/local/bin/task-compose \
    -ldflags "-s -w -X 'app.Version=${VERSION}' -X 'app.CommitHash=$(git rev-parse HEAD)' -X 'app.BuildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')' -X 'app.ExecutionMode=CLI' -X 'main.builtBy=goreleaser-docker'" \
    ./... # 編譯所有模組下的 Go 檔案

# --- 階段 3: 最終運行時影像 (Final Stage) ---
# 使用一個極其輕量級的基礎影像，只包含運行應用程式所需的最小依賴
FROM alpine/git:2.44.0 # 如果你的工具需要 git，這個基礎影像會包含

# 設定容器內的環境變數，如果需要
ENV PATH="/usr/local/bin:$PATH"

# 將編譯好的應用程式可執行檔從 builder 階段複製到最終影像中
COPY --from=builder /usr/local/bin/task-compose /usr/local/bin/task-compose

# 設置應用程式的入口點 (Entrypoint)
# 這會讓容器啟動時自動執行 task-compose
ENTRYPOINT ["task-compose"]

# 如果你的應用程式可以接受子命令作為預設行為，可以設置 CMD
# CMD ["help"]