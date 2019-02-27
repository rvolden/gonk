gonk :
	go build src/gonk.go

affine :
	go build src/affine.go

clean :
	rm -f gonk affine

install :
	cp gonk /usr/bin/
