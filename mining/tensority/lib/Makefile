CC = g++

ifeq ($(OS),Windows_NT)
    CCFLAGS += -D WIN32
    OS = win
    ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
        CCFLAGS += -D AMD64
        BITS = 64
    else
        ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
            CCFLAGS += -D AMD64
            BITS = 64
        endif
        ifeq ($(PROCESSOR_ARCHITECTURE),x86)
            CCFLAGS += -D IA32
            BITS = 32
        endif
    endif
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        CCFLAGS += -D LINUX
        OS = linux
    endif
    ifeq ($(UNAME_S),Darwin)
		CC=/usr/local/opt/llvm/bin/clang++
        CCFLAGS += -D OSX
        OS = darwin
    endif
    UNAME_P := $(shell uname -p)
    ifeq ($(UNAME_P),x86_64)
        CCFLAGS += -D AMD64
        BITS = 64  
    endif
    ifneq ($(filter %86,$(UNAME_P)),)
        CCFLAGS += -D IA32
        BITS = 32
    endif
    ifneq ($(filter arm%,$(UNAME_P)),)
        CCFLAGS += -D ARM
    endif
endif

.PHONY: clean

test: cSimdTs_$(OS)$(BITS).o
	cd .. && go test -v && go test -bench=.
	rm *.o

all: cSimdTs_$(OS)$(BITS).o

%_$(OS)$(BITS).o: %.cpp
	$(CC) -o $@ -c $^ -std=c++11 -pthread -mavx2 -O3 -fopenmp

clean:
	rm -f *.o *.so *.a
	