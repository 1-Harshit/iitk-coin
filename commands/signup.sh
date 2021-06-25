read -p "Enter Your Roll No: "  roll
read -p "Enter Your Name: "  name
read -p "Enter Your Email ID: "  email
read -p "Enter Your Batch "  batch
read -s -p "Enter Your Password: "  pass

curl --header "Content-Type: application/json"  --request POST  --data '{"roll":'$roll', "name":"'$name'", "email":"'$email'", "batch":"'$batch'", "password":"'$pass'"}'  http://localhost:8080/signup