read -p "Enter Your JW Token: "  JWT

curl -H "Authorization: Bearer $JWT"  http://localhost:8080/view