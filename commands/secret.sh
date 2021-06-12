read -p "Enter Your JW Token: "  JWT


curl -H "Authorization: Bearer $JWT" localhost:8080/secretpage
