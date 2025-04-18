package routes

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func GetUser(ctx *fasthttp.RequestCtx) {
	user := User{ID: 1, Name: "Alice"}
	response, _ := json.Marshal(user)
	ctx.SetContentType("application/json")
	ctx.SetBody(response)
}

func CreateUser(ctx *fasthttp.RequestCtx) {
	var user User
	if err := json.Unmarshal(ctx.PostBody(), &user); err != nil {
		ctx.Error("Invalid JSON", fasthttp.StatusBadRequest)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetContentType("application/json")
	ctx.SetBodyString(`{"message":"User created"}`)
}
