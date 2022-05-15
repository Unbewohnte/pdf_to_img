all:
	cd src/ && go build && mv pdf_to_img ..

windows:
	cd src/ && CC=/usr/bin/x86_64-w64-mingw32-gcc GOOS=windows CGO_ENABLED=1 GOARCH=amd64 go build -tags=static && mv pdf_to_img.exe ..