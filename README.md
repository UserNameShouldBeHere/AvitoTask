# AVITO SHOP [![License](https://img.shields.io/github/license/UserNameShouldBeHere/AvitoTask)](https://opensource.org/license/mit) [![Lang](https://img.shields.io/github/languages/top/UserNameShouldBeHere/SSO)](https://go.dev/)

## Задание
Ссылка на [задание](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/Backend/Backend-trainee-assignment-winter-2025/Backend-trainee-assignment-winter-2025.md)

## Проделанная работа
В качестве ЯП был выбран Go 1.23

БД PostgreSQL

В проекте настроен линтер, CI, система логирования

Для просмотра тестового покрытие можно воспользоваться командой `go test ./... -coverprofile='coverage.out' && cat coverage.out | grep -v 'mock' | grep -v 'proto' > coverage_cleaned.out go tool cover -func='coverage_cleaned.out'` или же просто `task test-cover`(для этого требуется установить [task](https://taskfile.dev/))

Для повышения скорости работы были добавлены индексы в postgres

Чтобы дополнительно улучить эффективность сервсиа, можно добавить кэширование(например, в Redis), а также очередь запросов, чтобы сервис накапливал запросы, которые не успевает обработать в отдельном хранилище, и не простаивал

## Инструкция по запуску
### Из консоли
Вначала необходимо заполнить базу данных с помощью [скрипта](db/init.sql)

Затем выполнить команду `go run cmd/app/main.go`

Чтобы указать пользователя и пароль от postgres можно воспользоваться флагами `-dbuser` и `-dbpass` соответственно

### Docker
Для запуска достаточно выполнить команду `docker-compose up -d`

### Podman
Необходимо собрать образы контейнеров с помощью команд `podman build -t app -f cmd/app/Dockerfile .` и `podman build -t postgres -f db/Dockerfile .`

Далее выполнить команду `podman kube play pod.yml`
