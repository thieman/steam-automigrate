UNAME_S := $(shell uname -s)

BIN := steam-automigrate.exe

run: windows
	./${BIN}

windows:
	GOOS=windows GOARCH=386 go build -o ${BIN} cmd/steam-automigrate/main.go

clean:
	rm -f ${BIN}

.PHONY: clean