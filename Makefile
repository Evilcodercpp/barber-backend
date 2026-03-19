.PHONY: run build test clean

# Запуск
run:
	go run cmd/server/main.go

# Сборка бинарника
build:
	go build -o bin/server cmd/server/main.go

# Создать БД
db-create:
	createdb barber

# Удалить БД
db-drop:
	dropdb barber

# Тесты
test:
	go test ./...

clean:
	rm -rf bin/
