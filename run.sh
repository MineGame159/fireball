cd build || exit

llc --filetype=obj -o=test.o test.ll
clang -lm -o test test.o

./test
echo Exit code: $?