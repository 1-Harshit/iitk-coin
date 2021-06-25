# iitk-coin
**Summer Project 2021**
---

### Task 4

- This is a web server with two endpoints `/login` and `/signup` that accepts POST requests.

- The `/signup` endpoint will receive a new `user`'s Roll no, Name, Email and password and create a new `user` in the database. Password is stored after hashing and salting by `bcrypt`. 

- The `/login` endpoint will take in the rollno and password and if verified successfully will return a JWT (JSON Web Token) as part of the response, valid for 5 mins. 

- Endpoint `/secretpage` that returns the User's Roll only if the user is logged in. (i.e. the JWT sent along the request should be a valid token and the user is authorized to access the endpoint. 

- Endpoint `/view` will return the amount of coins a user hold.

- Endpoint `/reward` is to reward some user some amount of coins

- Endpoint `/transfer` is for user to user tansfer of coins 

### How to use


1. Run the `main.go` file this will Listen and Serve on `localhost:8080`
2. Use `curl` to give input or you can use shell script files in `commands` :
   - run `bash ./commands/signup.sh` to register a new user. This will post on `localhost:8080/signup`
   - run `bash ./commands/signin.sh` to login a user.  This will post on `localhost:8080/login` and return a JWT valid for 5 Minutes.
   - run `bash ./commands/secret.sh` to view secretpage if you have JWT.  This will use `localhost:8080/secretpage`.
   - run `bash ./commands/coins.sh` to view amount of coins a user has.  This will use `localhost:8080/view`.
   - run `bash ./commands/reward.sh` to reward coins a user, i.e. creating money.  This will use `localhost:8080/reward`.
   - run `bash ./commands/transfer.sh` to transfer coins among users, provided the user owns that much coin.  This will use `localhost:8080/transfer`.

## How to post at endpoints
1. `SignUp`: Endpoint to signup  
    POST: Roll-int, Name-string,	Email-string, Password-string, Batch-int
2. `login`:  Endpoint to signin  
   POST: Roll-int, Password-string
3. `SecretPage`: Endpoint to verify Login  
   GET: Authentication Bearer Header
4. `reward`: Endpoint only accessible by Gensec AH  
   POST: Roll-int, Coins-int & Authentication Bearer Header
5. `transfer`: Endpoint accessible by all to transfer coins  
   POST: roll-int, coins-int & Authentication Bearer Header
6. `View`: Endpoint to get coins
   GET: Authentication Bearer Header

## about `config/settings.go`
   Some unkown variables are stored there.  
   
   - Path of DB
   - Max coins one can have
   - Minimum Events needed for transfer
   - tax
      - var IntraBatchTax float64 = 0.02
      - var InterBatchTax float64 = 0.33


Kthnxbye
---
