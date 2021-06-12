# iitk-coin
**Summer Project 2021**
---

### Task 2

- This is a web server with two endpoints `/login` and `/signup` that accepts POST requests.

- The `/signup` endpoint will receive a new `user`'s Roll no, Name, Email and password and create a new `user` in the database. Password is stored after hashing and salting by `bcrypt`. 

- The `/login` endpoint will take in the rollno and password and if verified successfully will return a JWT (JSON Web Token) as part of the response, valid for 5 mins. 

- Finally Endpoint `/secretpage` that returns some User's Name only if the user is logged in. (i.e. the JWT sent along the request should be a valid token and the user is authorized to access the endpoint. 

### How to use

1. Run the `main.go` file this will Listen and Serve on `localhost:8080`
2. Use `curl` to give input or you can use shell script files in `commands` :
   - run `bash ./commands/signup.sh` to register a new user. This will post on `localhost:8080/signup`
   - run `bash ./commands/signin.sh` to login a user.  This will post on `localhost:8080/login` and return a JWT valid for 5 Minutes.
   - run `bash ./commands/secret.sh` to View secretpage if you have JWT.  This will use `localhost:8080/secretpage`.

---
