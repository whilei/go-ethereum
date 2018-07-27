
# ./runtests-short.sh tests/

path="$1"

set GOMAXPROCS=6

noteid="$(git rev-parse HEAD).$(date +%Y-%m-%d-%H-%M-%S)"
echo $noteid

mkdir -p gotarchive

mv got*out* gotarchive/

echo "go test -v ./$path... &> got.$noteid.out"
go test -v ./$path... &> got.$noteid.out
echo "cat got.$noteid.out | grep FAIL | wc -l"
echo "PASS: $(cat got.$noteid.out | grep PASS | wc -l)"
echo "SKIP: $(cat got.$noteid.out | grep SKIP | wc -l)"
echo "FAIL: $(cat got.$noteid.out | grep FAIL | wc -l)"

echo "go test -tags=sputnikvm -v ./$path... &> got.$noteid.out.svm"
go test -tags=sputnikvm -v ./$path... &> got.$noteid.out.svm
echo "cat got.$noteid.out.svm | grep FAIL | wc -l"
echo "PASS: $(cat got.$noteid.out.svm | grep PASS | wc -l)"
echo "SKIP: $(cat got.$noteid.out.svm | grep SKIP | wc -l)"
echo "FAIL: $(cat got.$noteid.out.svm | grep FAIL | wc -l)"

unset GOMAXPROCS

./analyse-tests.sh "got.$noteid.out"
