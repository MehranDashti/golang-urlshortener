package handler

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
)

var validate = validator.New()

func bindAndValidate(c *gin.Context, req interface{}) *apperror.AppError {
    if err := c.ShouldBindJSON(req); err != nil {
        return apperror.BadRequest("invalid request body")
    }

    if err := validate.Struct(req); err != nil {
        errors := make(map[string]string)
        for _, e := range err.(validator.ValidationErrors) {
            errors[e.Field()] = validationMessage(e)
        }
        return apperror.BadRequestWithDetails("validation failed", errors)
    }

    return nil
}

func validationMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return e.Field() + " is required"
    case "url":
        return e.Field() + " must be a valid URL"
    case "email":
        return e.Field() + " must be a valid email"
    case "min":
        return e.Field() + " is too short (min " + e.Param() + ")"
    case "max":
        return e.Field() + " is too long (max " + e.Param() + ")"
    default:
        return e.Field() + " is invalid"
    }
}

func parsePagination(c *gin.Context) (model.PaginationParams, *apperror.AppError) {
    pageStr  := c.DefaultQuery("page", "1")
    limitStr := c.DefaultQuery("limit", "10")

    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        return model.PaginationParams{},
            apperror.BadRequest("page must be a positive integer")
    }

    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit < 1 || limit > 100 {
        return model.PaginationParams{},
            apperror.BadRequest("limit must be between 1 and 100")
    }

    return model.PaginationParams{Page: page, Limit: limit}, nil
}