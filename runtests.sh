set GOMAXPROCS=2

echo 'go test -v ./... &> gotest.out'
go test -v ./... &> gotest.out
echo 'go test -tags=sputnikvm -v ./... &> gotest.svm.out'
go test -tags=sputnikvm -v ./... &> gotest.svm.out
echo 'go test -v ./tests/... &> gotest_tests.out'
go test -v ./tests/... &> gotest_tests.out
echo 'go test -tags=sputnikvm -v ./tests/... &> gotest_tests.svm.out'
go test -tags=sputnikvm -v ./tests/... &> gotest_tests.svm.out

unset GOMAXPROCS
