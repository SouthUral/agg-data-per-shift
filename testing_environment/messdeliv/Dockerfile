#  указываем базовый образ, на основании которого будем делать сборку
FROM golang:1.21-alpine AS build

#  указываем рабочую директорию для проекта
WORKDIR /usr/local/src/

#  скачиваем зависимости, которые указаны в go.mod и в go.sum
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

#  копируем все файлы в рабочую директорию
COPY . .

#  собираем бинарник, на первом месте, это куда будет положен файл
#  относительный путь, на втором месте путь до main.go файла
RUN go build -o ./bin/messdeliv /usr/local/src/main.go

#  запуск бинарного файла
ENTRYPOINT ["/usr/local/src/bin/messdeliv"]