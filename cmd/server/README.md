# cmd/server

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение.

## Сборка с указанием версии

Для задания значений `buildVersion`, `buildDate` и `buildCommit` при компиляции используйте флаг `-ldflags`:

```bash
go build -ldflags "-X main.buildVersion=1.0.0 \
                  -X 'main.buildDate=$(date +%Y/%m/%d\ %H:%M:%S)' \
                  -X 'main.buildCommit=$(git rev-parse --short HEAD)'" \
         -o ./cmd/server ./cmd/server
```

Если флаги не заданы, при старте выводится `N/A`:

```
Build version: N/A
Build date: N/A
Build commit: N/A
```

С заданными значениями:

```
Build version: 1.0.0
Build date: 2026/06/25 07:40:00
Build commit: a1b2c3d
```
