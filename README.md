# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Профилирование и оптимизация

```bash
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

```
File: server
Build ID: bbb9d7a50e0efc5a1441beb4e462e1bcbb97043a
Type: inuse_space
Time: 2026-06-07 00:19:45 MSK
Showing nodes accounting for -9979.93kB, 47.98% of 20798.62kB total
      flat  flat%   sum%        cum   cum%
-6318.10kB 30.38% 30.38% -7927.84kB 38.12%  compress/flate.NewWriter (inline)
-1097.69kB  5.28% 35.66% -1097.69kB  5.28%  compress/flate.(*compressor).initDeflate (inline)
    -514kB  2.47% 38.13%     -514kB  2.47%  bufio.NewReaderSize (inline)
    -514kB  2.47% 40.60%     -514kB  2.47%  bufio.NewWriterSize (inline)
 -512.07kB  2.46% 43.06%  -512.07kB  2.46%  net/http.(*Server).newConn (inline)
 -512.05kB  2.46% 45.52%  -512.05kB  2.46%  compress/flate.newHuffmanBitWriter (inline)
 -512.01kB  2.46% 47.98%  -512.01kB  2.46%  net/textproto.(*Reader).ReadLine (inline)
         0     0% 47.98%     -514kB  2.47%  bufio.NewReader (inline)
         0     0% 47.98% -1609.74kB  7.74%  compress/flate.(*compressor).init
         0     0% 47.98% -7927.84kB 38.12%  compress/gzip.(*Writer).Write
         0     0% 47.98% -7927.84kB 38.12%  encoding/json.(*Encoder).Encode
         0     0% 47.98% -7927.84kB 38.12%  github.com/DimitryShR/go-ya-practicum-metrics/internal/compress.(*compressWriter).Write
         0     0% 47.98% -7927.84kB 38.12%  github.com/DimitryShR/go-ya-practicum-metrics/internal/handler.(*MetricHandler).UpdateMetricHandlerJSON
         0     0% 47.98% -7927.84kB 38.12%  github.com/DimitryShR/go-ya-practicum-metrics/internal/handler.writeJSON
         0     0% 47.98% -7927.84kB 38.12%  github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware.GzipMiddleware.func1
         0     0% 47.98% -7927.84kB 38.12%  github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware.LogRequest.func1
         0     0% 47.98% -7927.84kB 38.12%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 47.98% -7927.84kB 38.12%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 47.98% -7927.84kB 38.12%  github.com/go-chi/chi/v5/middleware.StripSlashes.func1
         0     0% 47.98% -7927.84kB 38.12%  main.newRouter.func1
         0     0% 47.98%  -512.07kB  2.46%  main.runServer.func1
         0     0% 47.98%  -512.07kB  2.46%  net/http.(*Server).ListenAndServe
         0     0% 47.98%  -512.07kB  2.46%  net/http.(*Server).Serve
         0     0% 47.98%  -512.01kB  2.46%  net/http.(*conn).readRequest
         0     0% 47.98% -9467.86kB 45.52%  net/http.(*conn).serve
         0     0% 47.98% -7927.84kB 38.12%  net/http.HandlerFunc.ServeHTTP
         0     0% 47.98%     -514kB  2.47%  net/http.newBufioReader
         0     0% 47.98%     -514kB  2.47%  net/http.newBufioWriterSize
         0     0% 47.98%  -512.01kB  2.46%  net/http.readRequest
         0     0% 47.98% -7927.84kB 38.12%  net/http.serverHandler.ServeHTTP
```
