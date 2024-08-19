# Chirps API Routes

## POST /api/chirps
#### Create Chirp
This is an authorized route meaning that it will look, for a specific header <code>Authorization: Bearer {JWT}</code>

Accepts a json body with less than or equal to 140 characters
<code>{
		body string
}</code>

Creates Chirp and adds to db.json

Success Response
<code>
{
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}
</code>

## GET /api/chirps
#### Get Chirps
##### Query Params
    ?sort=asc or ?sort=desc = sorting
    ?author_id={id} = get all chirps that belong to this author

Success Response
<code>[]{
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}</code>


## GET /api/chirps/{chirpID}
#### Get chirp by ID

Gets chirp based on id included in url
Success Response
<code>
{
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}
</code>

## DELETE /api/chirps/{chirpID}
#### Delete by ID

Deletes chirp based on id included in url
Success Response
<code>204: No Content</code>