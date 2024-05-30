# TaskPizSuWok
-----------------------------------------------------------------------------------------------------
Тестовое задание GO

Разработать сервис по получению информации по контейнеру - информация должна быть взята из Docker registry с тегами (документация https://istribution.github.io/distribution/spec/api/#listing-image-tags) с описанием количества слоев в контейнере и размер контейнера (документация по endpoint -
https://distribution.github.io/distribution/spec/api/#get-manifest |документация по manifest v2 schema
2(требуется только v2 без List) - https://distribution.github.io/distribution/spec/manifest-v2-2/)

На вход - имя контейнера postgres/postgres или postgres, как захотим.

На выход - JSON, словарь, где ключ это ссылка (к примеру postgres: 15), а значение - словарь состоящий из количества слоев и размера контейнера.

Требования к сервису:
 1 ﻿﻿﻿Должна быть система которая минимизирует запросы к Docker registry
 2 ﻿﻿﻿Должна быть система которая не даст слишком часто запрашивать по ір (значение будет
передаваться в header «X-Forwarded-For»)

Как развернуть Docker registry локально для тестов (https://distribution.github.io/distribution/)

-------------------------------------------------------------------------------------------------------
Для удовлетворения требованиям было сделано:
1 Кеширование для минимизации запросов к Docker registry и ускорения ответов
2 Лимит на обращение в виде ограничения запросов (1 запрос в 1 секунду) {при необходимости можно изменить, смотри комментарии в коде}
-------------------------------------------------------------------------------------------------------
Для проверки работоспособности необходимо:
1 Развернуть локальный Docker Registry:
docker run -d -p 5000:5000 --name registry registry:2
2 Загрузить тестовые образы в локальный registry:
 качаем образы:
 docker pull busybox:1.32
 docker pull nginx:latest
 docker pull alpine:latest
 docker pull redis:latest
 
 тегируем их для локал registry:
 docker tag busybox:1.32 localhost:5000/busybox:1.32
 docker tag nginx:latest localhost:5000/nginx:latest
 docker tag alpine:latest localhost:5000/alpine:latest
 docker tag redis:latest localhost:5000/redis:latest
 
 отправляем образы в локал registry:
 docker push localhost:5000/busybox:1.32
 docker push localhost:5000/nginx:latest
 docker push localhost:5000/alpine:latest
 docker push localhost:5000/redis:latest

3 Запускаем go run main.go (TaskPizSuWok/cmd)
 Проверяем работу. Например запросом: curl -H "X-Forwarded-For: 1.2.3.4" http://localhost:8080/busybox
 curl -H "X-Forwarded-For: 1.2.3.4" http://localhost:8080/nginx
 curl -H "X-Forwarded-For: 1.2.3.4" http://localhost:8080/alpine
 curl -H "X-Forwarded-For: 1.2.3.4" http://localhost:8080/redis
Либо же переходим на локалхост и смотрим страницу:
 http://localhost:8080/busybox
 http://localhost:8080/nginx
 http://localhost:8080/alpine
 http://localhost:8080/redis
Ожидаемый ответ: {"redis:latest":{"layers":8,"size":45511837}} (само название образа, его версия и кол-во слоев с размером)

-------------------------------------------------------------------------------------------------
Так же для простоты отслеживания ошибок было поставлено логирование. Таким образом можно отследить, к примеру, работу кеширования. При первом запросе в логах будет запись Cache miss for redis(образ, который проверяли). При повторном запросе будет уже Cache hit for redis, что подтвержает работу кеша. Если же изменить лимит на обращения с 1 секунды на запрос до 5 секунд, к примеру, то можно будет заметить как работает ограничение на количество запросов за заданное время. Ответ будет: Too many requests.
