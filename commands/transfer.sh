read -p "Enter sender Roll No: "  from
read -p "Enter Reciever Roll No: "  to
read -p "No of coins to be transfered: "  coins

curl --header "Content-Type: application/json"  --request POST  --data '{"from":'$from', "to":'$to', "coins":'$coins'}'  http://localhost:8080/transfer