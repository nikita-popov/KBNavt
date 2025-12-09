package api

import (
    "log/slog"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

// BasicAuthMiddleware provides HTTP Basic Authentication
func BasicAuthMiddleware(user, pass string, logger *slog.Logger) echo.MiddlewareFunc {
    return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
        if username == user && password == pass {
            return true, nil
        }
        logger.Warn("auth failed", "username", username)
        return false, nil
    })
}
