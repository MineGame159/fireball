cd build || exit

llc --filetype=obj -o=test.o test.ll
clang -o test test.o

./test
echo Exit code: $?