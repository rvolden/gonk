align :
	go build src/goAlign.go

clean :
	rm -f goAlign

install :
	cp goAlign /usr/bin/
