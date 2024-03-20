package route

import (
	"core/repo"
	"framework/node"
	"hall/handler"
)

func Register(r *repo.Manager) node.LogicHandler {
	handlers := make(node.LogicHandler)
	userHandler := handler.NewUserHandler(r)
	handlers["userHandler.updateUserAddress"] = userHandler.UpdateUserAddress
	return handlers
}
