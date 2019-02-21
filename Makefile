align :
	go build src/goAlign.go

affine :
	go build src/affine.go

clean :
	rm -f goAlign affine

install :
	cp goAlign /usr/bin/
