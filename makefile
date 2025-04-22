build:
	go build -o bin/shist

clean:
	rm -rf bin/

install-macos: build
	sudo cp bin/shist /usr/local/bin/shist
