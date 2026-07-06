# rsacrypto

Пакет `rsacrypto` предоставляет функции для загрузки RSA-ключей, а также для блочного шифрования и расшифровки данных с использованием RSA-OAEP-SHA256.

## Генерация ключей

### С помощью dev-утилиты

```bash
go run dev/gencerts/main.go
```

Будут созданы два файла в директории `.certs`:
- `cert.pem` — X.509 сертификат, содержащий публичный RSA-ключ (используется агентом)
- `private.pem` — приватный RSA-ключ в формате PKCS1 (используется сервером)

### С помощью OpenSSL

```bash
# 1. Генерация приватного ключа RSA-4096
openssl genrsa -out .certs/private.pem 4096

# 2. Создание самоподписанного сертификата
openssl req -x509 -new -nodes -key .certs/private.pem -sha256 -days 3650 \
  -subj "/C=RU/O=Yandex.Praktikum" \
  -addext "subjectAltName=IP:127.0.0.1,IP:::1" \
  -out .certs/cert.pem
```

## Запуск

### Агент с шифрованием

```bash
./agent -crypto-key .certs/cert.pem
# или через переменную окружения
CRYPTO_KEY=.certs/cert.pem ./agent
```

### Сервер с расшифровкой

```bash
./server -crypto-key .certs/private.pem
# или через переменную окружения
CRYPTO_KEY=.certs/private.pem ./server
```
