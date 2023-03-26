build: pre
	go build -o target .
pre:
	mkdir -p target
	cp -rf conf target
clean:
	rm -rf target/*
