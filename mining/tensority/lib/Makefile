ifeq ($(OS),Windows_NT)
    CCFLAGS += -D WIN32
    ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
        CCFLAGS += -D AMD64
    else
        ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
            CCFLAGS += -D AMD64
        endif
        ifeq ($(PROCESSOR_ARCHITECTURE),x86)
            CCFLAGS += -D IA32
        endif
    endif
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        CCFLAGS += -D LINUX
    endif
    ifeq ($(UNAME_S),Darwin)
        CCFLAGS += -D OSX
    endif
    UNAME_P := $(shell uname -p)
    ifeq ($(UNAME_P),x86_64)
        CCFLAGS += -D AMD64
    endif
    ifneq ($(filter %86,$(UNAME_P)),)
        CCFLAGS += -D IA32
    endif
    ifneq ($(filter arm%,$(UNAME_P)),)
        CCFLAGS += -D ARM
    endif
endif


.PHONY: clean

UNAME := $(shell uname -s)
ifeq ($(UNAME),Darwin)
	CC = /usr/local/opt/llvm/bin/clang++
else
	CC = g++
endif

test: cSimdTs.o
	cd .. && go test -v && go test -bench=.
	rm *.o

all: cSimdTs.o

%.o: %.cpp
	$(CC) -o $@ -c $^ -std=c++11 -pthread -mavx2 -O3 -fopenmp

clean:
	rm -f *.o *.so *.a
	