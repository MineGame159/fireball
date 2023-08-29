go build -o fireball fireball/cmd
fireball build test.fb

cd build || exit

llc --filetype=obj -o=test.o test.ll
clang -no-pie -lm -o test test.o

./test
echo Exit code: $?