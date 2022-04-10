#make top

PWD=`pwd`

run:
	go run cmd/topd/main.go

clean:
		rm -f *.o *.so topd


