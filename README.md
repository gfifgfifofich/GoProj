# build
```bash
chmod +x ./scripts/BuildAndRun.sh
./scripts/BuildAndRun.sh
```
Либо продублировать все команды с файла ./scripts/BuildAndRun.sh в консоли, в случае другой операционной системмы
# requered packages
Go, docker, dbmate
# misc
.env file contains 
```env
DATABASE_URL='postgres://postgres:qwerty@0.0.0.0:5436/postgres?sslmode=disable'
DB_PASSWORD=qwerty
SECRET = 'jwjnadfgh08yuegr0h0ubxcvasd'
```

для создания пользователя
post /auth/sign-up
body
{
    "name": "Testname1",
    "password": "Testpassword1"
}

для выдачи пары токенов (Хранятся в Cookies, refresh есть в бд как bcrypt hash):
post /task/access?guid=5c287eae-0520-4457-8efc-b53801716550

для обновления access токена:
post /task/refresh

