# chirpy

## Description
A web server constructed using golang including APIs for user creation, authorization, authentication as well as "chirps" which are tied to those users and require authentication and authorization to write.

## Why
This was solely a learning venture into golang and the construction of webservers using it.

## How to Use it
<ul>
    <li>Ensure you have golang installed I generally used v1.22.6</li>
    <li>perform a git clone of the project</li>
    <li>you should see a .env.example file you will need to include a .env with your own secret for this to work.</li>
    <li>to run it locally call <code>go build -o out && ./out</code> in your command line</li>
    <li>I have included a "debug" flag that will empty the json database and rebuild it for you if included <code>go build -o out && ./out --debug</code></li>
</ul>

Now it's up and running you can make get, put, post, and delete requests to the different apis.


## Where can I learn more
I recommend checking out [main.go](./main.go) or documentation on the [chirps api](./docs/chirps.md) or [users api](./docs/users.md)
