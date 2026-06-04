//go:build integration

package tests

import (
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "urlshortener/tests/testhelper"
    "urlshortener/tests/testserver"
)

func TestDeleteUser_DeletesLinksAndUser(t *testing.T) {
    s := testserver.New()
    defer s.CleanDB()

    // Create admin user
    adminToken := makeAdmin(t, s)

    // Create a regular user
    testhelper.MakeRequest(s.Router,
        http.MethodPost, "/api/v1/auth/signup",
        `{"email":"victim@test.com","password":"123456"}`, "")

    w := testhelper.MakeRequest(s.Router,
        http.MethodPost, "/api/v1/auth/login",
        `{"email":"victim@test.com","password":"123456"}`, "")

    data := testhelper.GetData(w)
    userToken := data["access_token"].(string)

    // Get user ID from list
    usersW := testhelper.MakeRequest(s.Router,
        http.MethodGet, "/api/v1/admin/users", "", adminToken)
    usersData := testhelper.GetData(usersW)
    users := usersData["data"].([]interface{})

    var userID string
    for _, u := range users {
        user := u.(map[string]interface{})
        if user["email"] == "victim@test.com" {
            userID = user["id"].(string)
        }
    }
    require.NotEmpty(t, userID)

    // Create some links as the victim user
    testhelper.MakeRequest(s.Router, http.MethodPost,
        "/api/v1/client/shorten",
        `{"url":"https://google.com"}`, userToken)
    testhelper.MakeRequest(s.Router, http.MethodPost,
        "/api/v1/client/shorten",
        `{"url":"https://github.com"}`, userToken)

    // Delete the user via admin
    deleteW := testhelper.MakeRequest(s.Router,
        http.MethodDelete,
        "/api/v1/admin/users/"+userID, "", adminToken)
    assert.Equal(t, http.StatusOK, deleteW.Code)

    // Verify user is gone
    usersW2 := testhelper.MakeRequest(s.Router,
        http.MethodGet, "/api/v1/admin/users", "", adminToken)
    usersData2 := testhelper.GetData(usersW2)
    users2 := usersData2["data"].([]interface{})

    for _, u := range users2 {
        user := u.(map[string]interface{})
        assert.NotEqual(t, "victim@test.com", user["email"],
            "deleted user should not appear in list")
    }

    // Verify links are gone
    linksW := testhelper.MakeRequest(s.Router,
        http.MethodGet, "/api/v1/admin/links", "", adminToken)
    linksData := testhelper.GetData(linksW)
    links := linksData["data"].([]interface{})

    for _, l := range links {
        link := l.(map[string]interface{})
        assert.NotEqual(t, userID, link["UserID"],
            "deleted user's links should not exist")
    }
}

// makeAdmin creates an admin user and returns their token
func makeAdmin(t *testing.T, s *testserver.TestServer) string {
    testhelper.MakeRequest(s.Router, http.MethodPost,
        "/api/v1/auth/signup",
        `{"email":"admin@test.com","password":"123456"}`, "")

    // Set admin role directly in DB
    s.DB.Exec("UPDATE users SET role = 'admin' WHERE email = 'admin@test.com'")

    w := testhelper.MakeRequest(s.Router, http.MethodPost,
        "/api/v1/auth/login",
        `{"email":"admin@test.com","password":"123456"}`, "")

    data := testhelper.GetData(w)
    return data["access_token"].(string)
}