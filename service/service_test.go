package service

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/SeijiOmi/user/db"
	"github.com/SeijiOmi/user/entity"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var client = new(http.Client)
var userDefault = entity.User{Name: "test", Email: "test@co.jp", Password: "password"}
var tmpBasePointURL string
var tmpBasePostURL string

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}

func setup() {
	tmpBasePointURL = os.Getenv("POINT_URL")
	tmpBasePostURL = os.Getenv("POST_URL")
	setTestURL()
	db.Init()
	initUserTable()
}

func teardown() {
	db.Close()
	os.Setenv("POINT_URL", tmpBasePointURL)
	os.Setenv("POST_URL", tmpBasePostURL)
}

func TestGetAll(t *testing.T) {
	initUserTable()
	createDefaultUser()
	createDefaultUser()

	var b Behavior
	users, err := b.GetAll()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(users))
}

func TestCreateModel(t *testing.T) {
	var b Behavior
	user, err := b.CreateModel(userDefault)

	assert.Equal(t, nil, err)
	assert.Equal(t, userDefault.Name, user.Name)
	assert.Equal(t, userDefault.Email, user.Email)
	assert.NotEqual(t, userDefault.Password, user.Password)
}

func TestGetByIDExists(t *testing.T) {
	user := createDefaultUser()
	var b Behavior
	user, err := b.GetByID(strconv.Itoa(int(user.ID)))

	assert.Equal(t, nil, err)
	assert.Equal(t, userDefault.Name, user.Name)
	assert.Equal(t, userDefault.Email, user.Email)
}

func TestGetByIDNotExists(t *testing.T) {
	var b Behavior
	user, err := b.GetByID(string(userDefault.ID))

	assert.NotEqual(t, nil, err)
	var nilUser entity.User
	assert.Equal(t, nilUser, user)
}

func TestUpdateByIDExists(t *testing.T) {
	user := createDefaultUser()

	updateUser := entity.User{Name: "not", Email: "not@co.jp", Password: "notpassword"}

	var b Behavior
	user, err := b.UpdateByID(strconv.Itoa(int(user.ID)), updateUser)

	assert.Equal(t, nil, err)
	assert.Equal(t, updateUser.Name, user.Name)
	assert.Equal(t, updateUser.Email, user.Email)
	assert.NotEqual(t, updateUser.Password, user.Password)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(updateUser.Password))
	assert.Equal(t, nil, err)
}

func TestUpdateByIDNotExists(t *testing.T) {
	user := createDefaultUser()

	updateUser := entity.User{Name: "not", Email: "not@co.jp", Password: "notpassword"}

	var b Behavior
	user, err := b.UpdateByID("0", updateUser)

	assert.NotEqual(t, nil, err)
	var nilUser entity.User
	assert.Equal(t, nilUser, user)
}

func TestDeleteByIDExists(t *testing.T) {
	initUserTable()
	user := createDefaultUser()

	db := db.GetDB()
	var beforeCount int
	db.Table("users").Count(&beforeCount)

	var b Behavior
	err := b.DeleteByID(strconv.Itoa(int(user.ID)))

	var afterCount int
	db.Table("users").Count(&afterCount)

	assert.Equal(t, nil, err)
	assert.Equal(t, beforeCount-1, afterCount)
}

func TestDeleteByIDNotExists(t *testing.T) {
	initUserTable()
	createDefaultUser()

	db := db.GetDB()
	var beforeCount int
	db.Table("users").Count(&beforeCount)

	var b Behavior
	err := b.DeleteByID("0")

	var afterCount int
	db.Table("users").Count(&afterCount)

	assert.Equal(t, nil, err)
	assert.Equal(t, beforeCount, afterCount)
}

func TestCreatePoint(t *testing.T) {
	err := createPoint(100,"testComment" , "testToken")
	assert.Equal(t, nil, err)
}

func TestCreatePointNotFoundErr(t *testing.T) {
	os.Setenv("POINT_URL", "http://unknown")
	err := createPoint(100,"testComment" , "testToken")
	assert.NotEqual(t, nil, err)
	setTestURL()
}

func TestTokenSuccess(t *testing.T) {
	user := userDefault
	token, err := createToken(user)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", token)

	id, err := perthToken(token)
	assert.Equal(t, nil, err)
	assert.Equal(t, user.ID, id)
}

func TestPerthTokenErr(t *testing.T) {
	id, err := perthToken("testToken")
	assert.NotEqual(t, nil, err)
	assert.Equal(t, uint(0), id)
}

func TestCreateHashPassword(t *testing.T) {
	hashPassword, err := createHashPassword(userDefault.Password)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, userDefault.Password, hashPassword)

	err = bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(userDefault.Password))
	assert.Equal(t, nil, err)
}

func TestTokenAuth(t *testing.T) {
	initUserTable()
	var b Behavior
	user := userDefault
	createUser, err := b.CreateModel(user)
	assert.Equal(t, nil, err)

	auth, err := b.LoginAuth(user.Email, user.Password)
	assert.Equal(t, nil, err)

	authUser, err := b.TokenAuth(auth.Token)
	assert.Equal(t, nil, err)
	assert.Equal(t, createUser, authUser)
}
func TestTokenAuthErr(t *testing.T) {
	initUserTable()

	user := userDefault
	token, err := createToken(user)
	assert.Equal(t, nil, err)

	var b Behavior
	authUser, err := b.TokenAuth(token)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, entity.User{}, authUser)
}

func TestLoginAuthUnknownUserErr(t *testing.T) {
	initUserTable()
	user := userDefault

	var b Behavior
	auth, err := b.LoginAuth(user.Email, user.Password)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, entity.Auth{}, auth)
}

func TestLoginAuthPasswordErr(t *testing.T) {
	initUserTable()
	user := userDefault

	var b Behavior
	_, err := b.CreateModel(user)
	assert.Equal(t, nil, err)
	auth, err := b.LoginAuth(user.Email, "unknownPassword")
	assert.NotEqual(t, nil, err)
	assert.Equal(t, entity.Auth{}, auth)
}

func TestGetMaxUserID(t *testing.T) {
	initUserTable()

	maxID, err := getMaxUserID()

	assert.Equal(t, nil, err)
	assert.Equal(t, 0, maxID)
}
func TestCreateUniqueDemouser(t *testing.T) {
	initUserTable()

	var b Behavior
	_, err := b.createUniqueDemoUser(userDefault)
	assert.Equal(t, nil, err)
	_, err = b.createUniqueDemoUser(userDefault)
	assert.Equal(t, nil, err)

	users, err := b.GetAll()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(users))
}
func TestCreateDemoData(t *testing.T) {
	initUserTable()

	var b Behavior
	demoAuth, err := b.CreateDemoData()
	assert.Equal(t, nil, err)

	_, err = b.TokenAuth(demoAuth.Token)
	assert.Equal(t, nil, err)

	users, err := b.GetAll()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(users))

	// 複数回デモユーザーを作成してもエラーとならない確認
	demoAuth, err = b.CreateDemoData()
	assert.Equal(t, nil, err)

	_, err = b.TokenAuth(demoAuth.Token)
	assert.Equal(t, nil, err)

	users, err = b.GetAll()
	assert.Equal(t, nil, err)
	assert.Equal(t, 4, len(users))
	fmt.Println(users)
}

func createDefaultUser() entity.User {
	db := db.GetDB()
	user := userDefault
	db.Create(&user)
	return user
}

func initUserTable() {
	db := db.GetDB()
	var u entity.User
	db.Delete(&u)
}

func setTestURL() {
	os.Setenv("POINT_URL", "http://user-mock-point:3000")
	os.Setenv("POST_URL", "http://user-mock-post:3000")
}
