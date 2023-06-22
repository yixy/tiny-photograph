build: pre
	go1.20.5 build -o target/tiny-photograph/tiny-photograph .
	tar -zcvf target/tiny-photograph.tar.gz -C target ./tiny-photograph
pre:
	go1.20.5 mod tidy
	mkdir -p target/tiny-photograph
	cp -rf conf target/tiny-photograph
clean:
	rm -rf target/*
