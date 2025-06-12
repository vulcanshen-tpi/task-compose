# 基礎映像：使用一個輕量級的 Linux 發行版。
# Alpine 是 Go 靜態編譯二進制文件的流行選擇，因為它非常小。
# 選擇一個穩定的版本，例如 alpine:3.18 或 alpine:3.19。
FROM alpine:3.18

# 安裝 ca-certificates：
# 如果你的 Go 應用程式會發送 HTTPS 請求 (例如呼叫外部 API)，
# 則需要安裝 ca-certificates 以確保 TLS/SSL 連接的正常工作。
# 如果你的應用程式完全不進行任何網路請求，這行可以省略。
RUN apk --no-cache add ca-certificates

# 設定容器內的工作目錄
# 所有後續的指令 (例如 COPY, ENTRYPOINT) 都會相對於這個目錄執行。
WORKDIR /app

# 複製由 GoReleaser 編譯好的主要二進制文件。
# GoReleaser 會在執行 Docker build 時，將這個二進制文件放在 Docker build context 的根目錄。
# 根據你的 builds.binary: "task-compose" 設定，複製後的檔案名就是 "task-compose"。
COPY task-compose .

# 複製任何你在 .goreleaser.yaml 的 dockers.extra_files 中指定的額外文件。
# 這些文件也會被 GoReleaser 放在 Docker build context 的根目錄。
COPY README.md .
COPY LICENSE .

# 設定容器啟動時的入口點。
# 這表示當容器運行時，它會直接執行 /app/task-compose。
# 使用 JSON 陣列格式是最佳實踐，避免 shell 處理。
ENTRYPOINT ["/app/task-compose"]

CMD ["-h"]

# 可選：如果你的應用程式是一個網路服務並監聽特定端口，可以暴露該端口。
# EXPOSE 8080