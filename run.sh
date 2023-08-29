cd build || exit

llc --filetype=obj -o=test.o test.ll
clang -no-pie -lm -o test test.o

./test
echo Exit code: $?