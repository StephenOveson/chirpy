# Users API Routes

## POST /api/users
#### Create User
This route accepts a json body in the format of 
<code>
    {
        "email": string
        "password": string
    }
</code>
<br />
This will generate a user in the db.json file found here [database](../database/db.json)
Password hashing is included in this route and is using bcrypt to do so.
<br />
Response:
<code>
    {
		Email        string `json:"email"`
		Id           int    `json:"id"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}
</code>

## POST /api/login
#### Login
This route also accepts a json body in the form of
<code>
    {
        "email": string
        "password": string
    }
</code>
<br />

This will perform a look up for the user comparing the hashed password and password passed in the body. 
It will generate a JWT for usage by further requests that require authorization see [put route](#put-apiusers)

<br />
Response:
<code>
    {
		Email        string `json:"email"`
		Id           int    `json:"id"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}
</code>
<br />

The token will be your JWT, and will last 1 hour, for requests that require authorization and the refresh token allows us to refresh our JSON Web Token when it expires, but the refresh token expires in 60 days see [refresh token](#post-apirefresh)


## PUT /api/users
#### Update User
This route accepts a json body of <code>{
    "email": string
    "password": string
}</code>
<br />
This will update the user using the included fields

This route also requires an Authorization header in the request in the form of <code>Authorization: "Bearer {JWT}"</code>
So your JSON Web token from logging in will be required.

## POST /api/refresh
#### Regenerate JWT
This accepts an authorization header of our refresh token to retrieve the JWT again once that has expired. The refresh token is included in the login response and can be included in your header in this format.
<code>Authorization: "Bearer {Refresh Token}"</code>
<br />
response: 
<code> 
    {
		Token string `json:"token"`
	}
</code>

## POST /api/revoke
#### Revoke Refresh token
This accepts an authorization header of our refresh token to retrieve the JWT again once that has expired. The refresh token is included in the login response and can be included in your header in this format.
<code>Authorization: "Bearer {Refresh Token}"</code>
<br />
response: No Content