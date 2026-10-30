package middleware

import "github.com/gin-gonic/gin"

func (p Principal) IsAdmin() bool   { return p.Role == RoleAdmin }
func (p Principal) IsTeacher() bool { return p.Role == RoleTeacher }
func (p Principal) IsStudent() bool { return p.Role == RoleStudent }

// Me возвращает текущего пользователя. Вызывать только в маршрутах за Auth.
func Me(c *gin.Context) Principal {
	p, _ := PrincipalFrom(c)
	return p
}
