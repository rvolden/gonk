align :
	go build src/goAlign.go

clean :
	rm -f goAlign SW_PARSE.txt

install :
	cp goAlign /usr/bin/
