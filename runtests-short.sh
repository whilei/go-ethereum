noteid="$1"

set GOMAXPROCS=2

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

echo 'go test -v ./tests/... &> gotest_tests.$noteid.out'
go test -v ./tests/... &> gotest_tests.$noteid.out
echo 'cat gotest_tests.$noteid.out | grep FAIL | wc -l'
cat gotest_tests.$noteid.out | grep FAIL | wc -l

echo 'go test -tags=sputnikvm -v ./tests/... &> gotest_tests.svm.$noteid.out'
go test -tags=sputnikvm -v ./tests/... &> gotest_tests.svm.$noteid.out
echo 'cat gotest_tests.svm.$noteid.out | grep FAIL | wc -l'
cat gotest_tests.svm.$noteid.out | grep FAIL | wc -l

unset GOMAXPROCS

