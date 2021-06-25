read -p "Enter Your Roll No: "  roll
read -p "No of coins to be awarded: "  coins
read -p "Enter Your JW Token: "  JWT

curl --header "Content-Type: application/json" -H "Authorization: Bearer $JWT" --request POST  --data '{"roll":'$roll', "coins":'$coins'}'  http://localhost:8080/reward