build: pre
	go build -o target .
	cd target
	tar -zcvf tiny-photograph.tar.gz -C target/ .
	mv tiny-photograph.tar.gz target/
pre:
	mkdir -p target
	cp -rf conf target
clean:
	rm -rf target/*
