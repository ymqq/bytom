TARGET = cSimdTs

ifeq ($(BITS),32)
	CC_PRE = i686
endif
ifeq ($(BITS),64)
	CC_PRE = x86_64
endif
ifeq ($(OS),win)
    CC_MI  = w64-mingw32
    CC_SUF = g++-posix
endif
ifeq ($(OS),linux)
    CC_MI  = linux-gnu
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
	
# clang++ -o linux32.o -c cSimdTs.cpp \
# -I/usr/i686-linux-gnu/include/c++/7/i686-linux-gnu/ \
# -std=c++11 -pthread -mavx2 -O3 -fopenmp=libomp -target i386-pc-linux-gnu -v


# clang++ -o linux64.o -c cSimdTs.cpp \
# -I/usr/include/x86_64-linux-gnu/c++/7/ \
# -std=c++11 -pthread -mavx2 -O3 -fopenmp=libomp -target x86_64-pc-linux-gnu -v

# clang++ -o win32.o -c cSimdTs.cpp \
# -I/usr/lib/gcc/i686-w64-mingw32/7.3-posix/include/c++/ \
# -I/usr/lib/gcc/i686-w64-mingw32/7.3-posix/include/c++/ \
# -std=c++11 -pthread -mavx2 -O3 -fopenmp=libomp -target i386-pc-windows-msvc -v


# clang++ -o win64.o -c cSimdTs.cpp \
# -I/usr/lib/gcc/x86_64-w64-mingw32/7.3-posix/include/c++/ \
# -std=c++11 -pthread -mavx2 -O3 -fopenmp=libomp -target x86_64-pc-windows-msvc -v
