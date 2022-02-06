# spammer

- Создать канал "Рассылка" в телеграме
- Создать файл .env с данными
- Запусить и следовать инструкциям
- Отправить сообщение в канал "Рассылка"

## how build

### windows

    go build -o .\bin\spammer.exe .\cmd\spammer\main.go

### linux / macos

    go build -o bin/spammer cmd/spammer/main.go

### .env

    APP_ID=id
    APP_HASH=hash
    USER_PHONE_NUMBER=+7-000-000-00-00
    CHATS=Чат 1,Чат 2
