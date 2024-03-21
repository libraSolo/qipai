package route

import (
	"core/repo"
	"framework/node"
	"game/handler"
	"game/logic"
)

func Register(r *repo.Manager) node.LogicHandler {
	handlers := make(node.LogicHandler)
	unionManager := logic.NewUnionManager()
	unionManager.CreateUnionById(1)
	unionHandler := handler.NewUnionHandler(r, unionManager)
	handlers["unionHandler.createRoom"] = unionHandler.CreateRoom
	return handlers
}
