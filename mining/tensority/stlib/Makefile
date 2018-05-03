.PHONY: clean

UNAME := $(shell uname -s)
ifeq ($(UNAME),Darwin)
	CC = /usr/local/opt/llvm/bin/clang++
else
	CC = g++
endif

test: cSimdTs.o
	rm -rf ../*.go
	cp *.go ../
	cd .. && go test -v && go test -bench=.
	rm *.o
	rm -rf ../*.go
	cp ../legacy/*.go ../

%.o: %.cpp
	$(CC) -o $@ -c $^ -std=c++11 -pthread -mavx2 -O3 -fopenmp

clean:
	rm -f *.o *.so *.a