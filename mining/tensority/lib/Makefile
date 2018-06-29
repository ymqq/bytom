TARGET = cSimdTs

ifeq ($(BITS),32)
	CC_PRE = i686
endif
ifeq ($(BITS),64)
	CC_PRE = x86_64
endif
ifeq ($(OS),win)
    CC_MI = w64-mingw32
    CC_SUF = g++-posix
endif
ifeq ($(OS),linux)
    CC_MI = linux-gnu
    CC_SUF = g++
endif

CCFLAGS += -std=c++11 -pthread -mavx2 -O3 -fopenmp

.PHONY: clean

all:
	OS=win 	   BITS=32 make $(TARGET)
	OS=win     BITS=64 make $(TARGET)
	OS=linux   BITS=32 make $(TARGET)
	OS=linux   BITS=64 make $(TARGET)

$(TARGET): $(TARGET).cpp
	$(CC_PRE)-$(CC_MI)-$(CC_SUF) -o $@_$(OS)$(BITS).o -c $^ $(CCFLAGS)

clean:
	rm -f *.o *.so *.a
	