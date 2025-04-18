### перед тем как запустить нужно создать .env файл в папке token_service, вот такого вида

```ini
HTTPServer=localhost:8080
POSTGRES_USER=postgres
POSTGRES_PASSWORD=321SuperPassword123
POSTGRES_DB=token_auth
STORAGE_PATH=postgres://postgres:321SuperPassword123@db:5432/token_auth?sslmode=disable
AccessTTL=15m
RefreshTTL=720h
CookieTTL=720h
HMACSecret=lkOjCCkZren/ynA/LP3DJd8AgCy/2CQdAtKiemT1JvdlbYMXakaZaYq0LX+peTillCPVJsksLDpjPLfh4NnaLA==
```

### чтобы создать HMACSecret можно использовать команду ``` openssl rand -base64 64 ```

## запуск ``` docker-compose up ```