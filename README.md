# build
```bash
chmod +x ./scripts/BuildAndRun.sh
./scripts/BuildAndRun.sh
```
Либо продублировать все команды с файла ./scripts/BuildAndRun.sh в консоли, в случае другой операционной системмы
# requered packages
Go, docker, dbmate (для создания, но в бд одна таблица, можно просто и завести запрос из db/migrations/20240907070455_DB.sql)
# misc
.env file contains 
```env
DATABASE_URL='postgres://postgres:qwerty@0.0.0.0:5436/postgres?sslmode=disable'
DB_PASSWORD=qwerty
SECRET = 'jwjnadfgh08yuegr0h0ubxcvasd'
```
DATABASE_URL используется в dbmate, SECRET - jwtkey

для создания пользователя
post /auth/sign-up
body
{
    "email": "Testmail@example.com",
    "password": "Testpassword1"
}

для выдачи пары токенов (Хранятся в Cookies, refresh есть в бд как bcrypt hash):
post /task/access?guid=5c287eae-0520-4457-8efc-b53801716550

для обновления access токена:
post /task/refresh

при использовании refresh с другим айпишником, будет роизводиться попытка отправки email на почту пользователя

начальная точка - cmd/main.go

Всё что касается обработки запросов и прочего с интернетов - в папке handler

Всё что касается внутренней обработки (создание токенов, хеширование и прочее) - в папке services

Всё что касается общения с бд - в папке repository

