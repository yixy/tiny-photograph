build: pre
	go1.20 build -o target .
	cd target
	tar -zcvf tiny-photograph.tar.gz -C target/ .
	mv tiny-photograph.tar.gz target/
pre:
	go1.20 mod tidy
	mkdir -p target
	cp -rf conf target
clean:
	rm -rf target/*
