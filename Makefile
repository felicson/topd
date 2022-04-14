#make top

PWD=`pwd`

run:
	go run cmd/topd/*.go

clean:
		rm -f *.o *.so topd


