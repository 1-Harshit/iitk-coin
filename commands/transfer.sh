read -p "Enter Reciever Roll No: "  to
read -p "No of coins to be transfered: "  coins
read -p "Enter Your JW Token: "  JWT


curl --header "Content-Type: application/json" -H "Authorization: Bearer $JWT" --request POST  --data '{"roll":'$to', "coins":'$coins'}'  http://localhost:8080/transfer