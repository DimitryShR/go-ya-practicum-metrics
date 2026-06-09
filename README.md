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

Узкие места выявлялись с помощью `pprof` (профили памяти «до» `profiles/base.pprof` и  «после» `profiles/result.pprof`)

Дополнительные замеры осуществлялись с помощью бенчмарков (результаты «до» `profiles/bench_base.txt`, «после» `profiles/bench_result.txt`)

### Выполненные оптимизации

1. **Переиспользование gzip.Writer/Reader через `sync.Pool`** (`internal/compress/gzip.go`):
   На каждый HTTP-запрос создавались новые `gzip.Writer`, `gzip.Reader` и `bytes.Buffer`. `compress/flate.NewWriter` занимал 15,3 MB flat (73,8% от 20,8 MB общего потребления). После внедрения трёх `sync.Pool` его потребление снизилось до 0,9 MB, а общее потребление памяти сервера упало до 2,7 MB (сокращение на 87%). Бенчмарки подтверждают: аллокации — с 20 → 1 alloc/op, память — с ~814 KB → 80–479 B (Small/Large payload), скорость — в 5–12 раз быстрее.

2. **Кеширование HTML-шаблона** (`internal/handler/metrichandler.go`):
   При каждом `GET /` HTML-шаблон парсился заново вместе с `regexp.Compile`. Шаблон вынесен в глобальную переменную `metricsTemplate` с однократной инициализацией. Результат по бенчмарку `GetAllMetrics`: аллокации — с 752 → 555 alloc/op (−26%), память — с 40,6 → 22,4 KB (−45%), время — со 159 → 100 μs (−37%).

3. **Предварительное выделение слайса метрик** (`internal/agent/metriccollector.go`):
   Создание слайса без ёмкости (`var metrics []models.Metrics`) приводило к множественным переаллокациям при `append`. Замена на `make(..., len(c.metrics)+1)` дала по бенчмарку `GetMetricsForReport`: аллокации — с 38 → 33 alloc/op (−13%), память — с 4,8 → 2,4 KB (−50%), время — с 7,4 → 5,6 μs (−25%).



Итоговый `pprof diff_base` подтверждает: суммарное сокращение потребления памяти составило **87%** (с 20,8 MB до 2,7 MB).

```bash
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

```
File: server
Build ID: 422d36284bb6b81297d51aa0e98d1b19139330c5
Type: inuse_space
Time: 2026-06-07 00:19:45 MSK
Showing nodes accounting for -18101.61kB, 87.03% of 20798.62kB total
      flat  flat%   sum%        cum   cum%
-14441.38kB 69.43% 69.43% -16051.12kB 77.17%  compress/flate.NewWriter (inline)
-1097.69kB  5.28% 74.71% -1097.69kB  5.28%  compress/flate.(*compressor).initDeflate (inline)
    -514kB  2.47% 77.18%     -514kB  2.47%  bufio.NewReaderSize (inline)
 -512.34kB  2.46% 79.65%  -512.34kB  2.46%  regexp/syntax.(*compiler).inst (inline)
 -512.23kB  2.46% 82.11%  -512.23kB  2.46%  runtime.mallocgc
  512.17kB  2.46% 79.65%   512.17kB  2.46%  net/http.Header.Clone (inline)
 -512.07kB  2.46% 82.11%  -512.07kB  2.46%  net/http.(*Server).newConn (inline)
 -512.05kB  2.46% 84.57%  -512.05kB  2.46%  compress/flate.newHuffmanBitWriter (inline)
 -512.01kB  2.46% 87.03%  -512.01kB  2.46%  net/textproto.(*Reader).ReadLine (inline)
         0     0% 87.03%     -514kB  2.47%  bufio.NewReader (inline)
         0     0% 87.03% -1609.74kB  7.74%  compress/flate.(*compressor).init
         0     0% 87.03% -16051.12kB 77.17%  compress/gzip.(*Writer).Write
         0     0% 87.03% -16051.12kB 77.17%  encoding/json.(*Encoder).Encode
         0     0% 87.03% -16051.12kB 77.17%  github.com/DimitryShR/go-ya-practicum-metrics/internal/compress.(*compressWriter).Write
         0     0% 87.03%   512.17kB  2.46%  github.com/DimitryShR/go-ya-practicum-metrics/internal/compress.(*compressWriter).WriteHeader
         0     0% 87.03% -15538.95kB 74.71%  github.com/DimitryShR/go-ya-practicum-metrics/internal/handler.(*MetricHandler).UpdateMetricHandlerJSON
         0     0% 87.03% -15538.95kB 74.71%  github.com/DimitryShR/go-ya-practicum-metrics/internal/handler.writeJSON
         0     0% 87.03%   512.17kB  2.46%  github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware.(*loggingResponseWriter).WriteHeader
         0     0% 87.03% -15538.95kB 74.71%  github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware.GzipMiddleware.func1
         0     0% 87.03% -15538.95kB 74.71%  github.com/DimitryShR/go-ya-practicum-metrics/internal/middleware.LogRequest.func1
         0     0% 87.03% -15538.95kB 74.71%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 87.03% -15538.95kB 74.71%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 87.03% -15538.95kB 74.71%  github.com/go-chi/chi/v5/middleware.StripSlashes.func1
         0     0% 87.03%  -512.34kB  2.46%  html/template.init
         0     0% 87.03% -15538.95kB 74.71%  main.newRouter.func1
         0     0% 87.03%  -512.07kB  2.46%  main.runServer.func1
         0     0% 87.03%  -512.07kB  2.46%  net/http.(*Server).ListenAndServe
         0     0% 87.03%  -512.07kB  2.46%  net/http.(*Server).Serve
         0     0% 87.03%  -512.01kB  2.46%  net/http.(*conn).readRequest
         0     0% 87.03% -16564.96kB 79.64%  net/http.(*conn).serve
         0     0% 87.03%   512.17kB  2.46%  net/http.(*response).WriteHeader
         0     0% 87.03% -15538.95kB 74.71%  net/http.HandlerFunc.ServeHTTP
         0     0% 87.03%     -514kB  2.47%  net/http.newBufioReader
         0     0% 87.03%  -512.01kB  2.46%  net/http.readRequest
         0     0% 87.03% -15538.95kB 74.71%  net/http.serverHandler.ServeHTTP
         0     0% 87.03%  -512.34kB  2.46%  regexp.Compile (inline)
         0     0% 87.03%  -512.34kB  2.46%  regexp.MustCompile
         0     0% 87.03%  -512.34kB  2.46%  regexp.compile
         0     0% 87.03%  -512.34kB  2.46%  regexp/syntax.(*compiler).compile
         0     0% 87.03%  -512.34kB  2.46%  regexp/syntax.(*compiler).rune
         0     0% 87.03%  -512.34kB  2.46%  regexp/syntax.Compile
         0     0% 87.03%  -512.34kB  2.46%  runtime.doInit (inline)
         0     0% 87.03%  -512.34kB  2.46%  runtime.doInit1
         0     0% 87.03%  -512.34kB  2.46%  runtime.main
         0     0% 87.03%  -512.23kB  2.46%  runtime.malg
         0     0% 87.03%  -512.23kB  2.46%  runtime.newobject
         0     0% 87.03%  -512.23kB  2.46%  runtime.newproc.func1
         0     0% 87.03%  -512.23kB  2.46%  runtime.newproc1
         0     0% 87.03%  -512.23kB  2.46%  runtime.systemstack
```
