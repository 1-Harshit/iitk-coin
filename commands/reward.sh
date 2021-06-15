read -p "Enter Your Roll No: "  roll
read -p "No of coins to be awarded: "  coins

curl --header "Content-Type: application/json"  --request POST  --data '{"roll":'$roll', "coins":'$coins'}'  http://localhost:8080/reward