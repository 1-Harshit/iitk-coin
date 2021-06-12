read -p "Enter Your Roll No: "  roll
read -s -p "Enter Your Password: "  pass

curl --header "Content-Type: application/json"  --request POST  --data '{"roll":'$roll', "password":"'$pass'"}'  http://localhost:8080/login