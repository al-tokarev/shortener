# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
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

## Профилирование (инкремент 17)

Наиболее узким местом в base.pprof оказался локальный репозиторий: при каждом добавлении открывался файл с хранилищем url, а также блокировалось мьютексом добавление url в память в методах saveLocal (лишний мьютекс, тк проще было добавлять в память в основном методе). Было сделано: открытие файла стало один раз на всю работу сервиса, убраны методы saveLocal. И также по аналогии с массовым добавлением batch. В результате затрачиваемая память существенно сократилась (в большей степени из-за единоразового открытия файла с хранилищем)

alexandr@MacBook-Pro-Aleksandr shortener % go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
File: main
Type: inuse_space
Time: 2026-08-19 16:43:40 MSK
Showing nodes accounting for -1077.30kB, 25.94% of 4152.46kB total
Dropped 2 nodes (cum <= 20.76kB)
      flat  flat%   sum%        cum   cum%
 -557.26kB 13.42% 13.42%  -557.26kB 13.42%  github.com/al-tokarev/shortener/internal/repository/urlrepository.(*LocalRepository).saveLocal
 -520.04kB 12.52% 25.94%  -520.04kB 12.52%  go.uber.org/zap/buffer.(*Buffer).Write (inline)
         0     0% 25.94%  -557.26kB 13.42%  github.com/al-tokarev/shortener/internal/auth.AuthMiddleware.func1
         0     0% 25.94%  -557.26kB 13.42%  github.com/al-tokarev/shortener/internal/handler/urlhandlers.(*Handler).GetShortenedUrl
         0     0% 25.94%  -557.26kB 13.42%  github.com/al-tokarev/shortener/internal/repository/urlrepository.(*LocalRepository).Save
         0     0% 25.94% -1077.30kB 25.94%  github.com/al-tokarev/shortener/internal/router.NewRouter.GzipMiddleware.func2.1
         0     0% 25.94% -1077.30kB 25.94%  github.com/al-tokarev/shortener/internal/router.NewRouter.WithLogging.func1.1
         0     0% 25.94%  -557.26kB 13.42%  github.com/al-tokarev/shortener/internal/service/urlservices.(*Service).SetUrl
         0     0% 25.94% -1077.30kB 25.94%  github.com/go-chi/chi.(*Mux).ServeHTTP
         0     0% 25.94%  -557.26kB 13.42%  github.com/go-chi/chi.(*Mux).routeHTTP
         0     0% 25.94%  -520.04kB 12.52%  go.uber.org/zap.(*Logger).With
         0     0% 25.94%  -520.04kB 12.52%  go.uber.org/zap.(*SugaredLogger).With
         0     0% 25.94%  -520.04kB 12.52%  go.uber.org/zap/zapcore.(*ioCore).With
         0     0% 25.94%  -520.04kB 12.52%  go.uber.org/zap/zapcore.(*ioCore).clone (inline)
         0     0% 25.94%  -520.04kB 12.52%  go.uber.org/zap/zapcore.(*jsonEncoder).Clone
         0     0% 25.94%  -520.04kB 12.52%  go.uber.org/zap/zapcore.consoleEncoder.Clone
         0     0% 25.94% -1077.30kB 25.94%  net/http.(*conn).serve
         0     0% 25.94% -1077.30kB 25.94%  net/http.HandlerFunc.ServeHTTP
         0     0% 25.94% -1077.30kB 25.94%  net/http.serverHandler.ServeHTTP
         0     0% 25.94%     -513kB 12.35%  runtime.allocm
         0     0% 25.94%   512.23kB 12.34%  runtime.malg
         0     0% 25.94%     -513kB 12.35%  runtime.mcall
         0     0% 25.94%     -513kB 12.35%  runtime.newm
         0     0% 25.94%   512.23kB 12.34%  runtime.newproc.func1
         0     0% 25.94%   512.23kB 12.34%  runtime.newproc1
         0     0% 25.94%     -513kB 12.35%  runtime.park_m
         0     0% 25.94%     -513kB 12.35%  runtime.resetspinning
         0     0% 25.94%     -513kB 12.35%  runtime.schedule
         0     0% 25.94%     -513kB 12.35%  runtime.startm
         0     0% 25.94%   512.23kB 12.34%  runtime.systemstack
         0     0% 25.94%     -513kB 12.35%  runtime.wakep