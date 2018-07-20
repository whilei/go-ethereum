path="$1"

set GOMAXPROCS=6

noteid="$(git rev-parse HEAD).$(date +%Y-%m-%d-%H-%M-%S)"
echo $noteid

# echo 'go test -v ./... &> gotest.$noteid.out'
# go test -v ./... &> gotest.$noteid.out
# echo 'cat gotest.$noteid.out | grep FAIL | wc -l'
# cat gotest.$noteid.out | grep FAIL | wc -l
# 
# echo 'go test -tags=sputnikvm -v ./... &> gotest.svm.$noteid.out'
# go test -tags=sputnikvm -v ./... &> gotest.svm.$noteid.out
# echo 'cat gotest.svm.$noteid.out | grep FAIL | wc -l'
# cat gotest.svm.$noteid.out | grep FAIL | wc -l

echo 'go test -v ./$path... &> got.$noteid.out'
go test -v ./$path... &> got.$noteid.out
echo 'cat got.$noteid.out | grep FAIL | wc -l'
cat got.$noteid.out | grep FAIL | wc -l

echo 'go test -tags=sputnikvm -v ./$path... &> got.svm.$noteid.out'
go test -tags=sputnikvm -v ./$path... &> got.svm.$noteid.out
echo 'cat got.svm.$noteid.out | grep FAIL | wc -l'
cat got.svm.$noteid.out | grep FAIL | wc -l

unset GOMAXPROCS

