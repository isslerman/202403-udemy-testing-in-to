Program that check is a number is a prime or not. 
There is a complete test file that check all the functions that you can test in this project. 

To run all tests run:
go test ./... -v

To check if the test tests all the funcions you have the alias:
gtcover

that runs: 
go test -coverprofile=coverage.out && go tool cover -html=coverage.out

others tests commands:
go test -v -run Test_alpha 
go test . 
