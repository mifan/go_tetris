package types

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"
)

var levels = map[int]string{}
var errUserNotExist = "User %v does not exist"

// Users is a cache which has an expire(in days)
type Users struct {
	// index by userId, email
	users     map[int]*User
	emails    map[string]*User
	nicknames map[string]*User
	mu        sync.RWMutex

	// user who is playing or observing a game
	busyUsers map[int]int64
	bmu       sync.RWMutex

	// generate next user Id
	nextId int
	nmu    sync.Mutex
}

func NewUsers() *Users {
	us := &Users{
		users:     make(map[int]*User),
		emails:    make(map[string]*User),
		nicknames: make(map[string]*User),
		busyUsers: make(map[int]int64),
	}
	return us.init()
}

// init users cache
func (us *Users) init() *Users {
	return us
}

// set next id
func (us *Users) SetNextId(val int) {
	us.nmu.Lock()
	defer us.nmu.Unlock()
	us.nextId = val
}

// incr next id
func (us *Users) IncrNextId() {
	us.SetNextId(us.GetNextId() + 1)
}

// get next id
func (us *Users) GetNextId() int {
	us.nmu.Lock()
	defer us.nmu.Unlock()
	return us.nextId
}

var errAlreadyExist = "User %v is already exist"

// add new users into the cache
func (us *Users) Add(users ...*User) error {
	us.mu.Lock()
	defer us.mu.Unlock()
	// first check if the uid, email is already exist
	// if exist, return errAlreadyExist
	// else insert and return nil
	for _, u := range users {
		if us.users[u.Uid] != nil {
			return fmt.Errorf(errAlreadyExist, u.Uid)
		}
		if us.emails[u.Email] != nil {
			return fmt.Errorf(errAlreadyExist, u.Email)
		}
		if us.nicknames[u.Nickname] != nil {
			return fmt.Errorf(errAlreadyExist, u.Nickname)
		}
		us.users[u.Uid] = u
		us.emails[u.Email] = u
		us.nicknames[u.Nickname] = u
	}
	return nil
}

// if email exist
func (us *Users) IsEmailExist(email string) bool {
	return us.GetByEmail(email) != nil
}

// is nickname exist
func (us *Users) IsNicknameExist(nickname string) bool {
	return us.GetByNickname(nickname) != nil
}

// delete users from cache
func (us *Users) del(uid ...int) {
	us.mu.Lock()
	defer us.mu.Unlock()
	for _, v := range uid {
		u, ok := us.users[v]
		if ok {
			delete(us.emails, u.Email)
			delete(us.nicknames, u.Nickname)
			delete(us.users, v)
		}
	}
}

// get a user
func (us *Users) GetById(uid int) *User {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.users[uid]
}

func (us *Users) GetByEmail(email string) *User {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.emails[email]
}

func (us *Users) GetByNickname(nickname string) *User {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.nicknames[nickname]
}

// set users in busy mode
func (us *Users) SetBusy(uids ...int) {
	us.bmu.Lock()
	defer us.bmu.RUnlock()
	t := time.Now().Unix()
	for _, uid := range uids {
		us.busyUsers[uid] = t
	}
}

// set users in free mode
func (us *Users) SetFree(uids ...int) {
	us.bmu.Lock()
	defer us.bmu.Unlock()
	for _, uid := range uids {
		delete(us.busyUsers, uid)
	}
}

// check if user is in busy mode
func (us *Users) IsBusyUser(uid int) bool {
	us.bmu.RLock()
	defer us.bmu.RUnlock()
	return us.busyUsers[uid] == 0
}

// interface for updating user information
type UpdateInterface interface {
	Field() string
	Val() interface{}
}

// update string field
type updateString struct {
	field string
	val   string
}

func NewUpdateString(field string, val string) updateString {
	return updateString{
		field: field,
		val:   val,
	}
}

func (us updateString) Field() string {
	return us.field
}

func (us updateString) Val() interface{} {
	return us.val
}

// update int field
type updateInt struct {
	field string
	val   int
}

func NewUpdateInt(field string, val int) updateInt {
	return updateInt{
		field: field,
		val:   val,
	}
}

func (ui updateInt) Field() string {
	return ui.field
}

func (ui updateInt) Val() interface{} {
	return ui.val
}

// update []byte field
type update2dByte struct {
	field string
	val   []byte
}

func NewUpdate2dByte(field string, val []byte) update2dByte {
	return update2dByte{
		field: field,
		val:   val,
	}
}

func (u2b update2dByte) Field() string {
	return u2b.field
}

func (u2b update2dByte) Val() interface{} {
	return u2b.val
}

// update
func (us *Users) Update(uid int, upts ...UpdateInterface) error {
	us.mu.Lock()
	defer us.mu.Unlock()
	if len(upts) == 0 {
		return nil
	}
	u, ok := us.users[uid]
	if !ok {
		return fmt.Errorf(errUserNotExist, uid)
	}
	return u.Update(upts...)
}

// user fields which could be updated
const (
	errCantUpdateField = "Can not update the user field %v"
	UF_Nickname        = "Nickname"
	UF_Avatar          = "Avatar"
	UF_Password        = "Password"
	UF_Energy          = "Energy"
	UF_Level           = "Level"
	UF_Win             = "Win"
	UF_Lose            = "Lose"
	UF_Addr            = "Addr"
	UF_Balance         = "Balance"
	UF_Freezed         = "Freezed"
	UF_Updated         = "Updated"
)

// initialize user field cache
func init() {
	u := User{}
	typ := reflect.TypeOf(u)
	val := reflect.ValueOf(u)
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).CanInterface() && typ.Field(i).Tag.Get("fixed") != "true" {
			userFields[typ.Field(i).Name] = true
		}
	}
}

var userFields = make(map[string]bool)

func canUserFieldUpdate(field string) bool {
	return userFields[field]
}

// fixed tag means it is not going to update the field by reflect method
type User struct {
	Uid      int `fixed:"true"`
	Avatar   []byte
	Email    string `fixed:"true"`
	Password string
	Nickname string `fixed:"true"`
	Energy   int
	Level    int
	Win      int
	Lose     int
	Addr     string `fixed:"true"`
	Balance  int
	Freezed  int
	Updated  int
	conn     *net.TCPConn
	mu       sync.Mutex
}

func (u User) String() string {
	b, _ := json.Marshal(u)
	return string(b)
}

func NewUser(uid int, email, password, nickname, addr string) *User {
	return &User{
		Uid:      uid,
		Email:    email,
		Password: password,
		Nickname: nickname,
		Addr:     addr,
		Updated:  int(time.Now().Unix()),
	}
}

var ErrNilConn = fmt.Errorf("the tcp connection is nil")

// get uid
func (u *User) GetUid() int {
	if u == nil {
		return -1
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Uid
}

// get current energy
func (u *User) GetEnergy() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Energy
}

// get current balance
func (u *User) GetBalance() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Balance
}

// get current freezed
func (u *User) GetFreezed() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.Freezed
}

// update user
func (u *User) Update(upts ...UpdateInterface) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	// check if field really exist
	for _, v := range upts {
		if !canUserFieldUpdate(v.Field()) {
			return fmt.Errorf(errCantUpdateField, v.Field())
		}
	}
	// update it
	u.Updated = int(time.Now().Unix())
	for _, v := range upts {
		reflect.Indirect(reflect.ValueOf(u)).FieldByName(v.Field()).Set(reflect.ValueOf(v.Val()))
	}
	return nil
}

// set tcp connection
func (this *User) SetConn(conn *net.TCPConn) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.conn = conn
}

// get tcp connection
func (this *User) GetConn() *net.TCPConn {
	if this == nil {
		return nil
	}
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.conn
}

// set read timeout in secs
func (this *User) SetReadTimeoutInSecs(n int) error {
	if c := this.GetConn(); c != nil {
		return c.SetReadDeadline(time.Now().Add(time.Duration(n) * time.Second))
	}
	return ErrNilConn
}

// set write timeout in secs
func (this *User) SetWriteTimeoutInSecs(n int) error {
	if c := this.GetConn(); c != nil {
		return c.SetWriteDeadline(time.Now().Add(time.Duration(n) * time.Second))
	}
	return ErrNilConn
}

// close
func (this *User) Close() error {
	if c := this.GetConn(); c != nil {
		return c.Close()
	}
	return ErrNilConn
}

// marshal the user
func (this User) MarshalJSON() ([]byte, error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	return json.Marshal(map[string]interface{}{
		"id":       this.Uid,
		"email":    this.Email,
		"nickname": this.Nickname,
		"level":    this.Level,
		"win":      this.Win,
		"lose":     this.Lose,
		"addr":     this.Addr,
		"balance":  this.Balance,
		"updated":  this.Updated,
	})
}

// sql generator
// update or insert
func (this *User) SqlGeneratorUpdate() (sql string, args []interface{}) {
	this.mu.Lock()
	defer this.mu.Unlock()
	typ := reflect.TypeOf(this).Elem()
	val := reflect.Indirect(reflect.ValueOf(this))
	updates := make([]string, 0)
	args = make([]interface{}, 0)
	sql = "INSERT INTO users("
	var l int
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).CanInterface() {
			if i != 0 {
				sql += ", "
			}

			l++
			f := typ.Field(i).Name
			if canUserFieldUpdate(f) {
				updates = append(updates, f)
			}
			sql += f
			args = append(args, val.Field(i).Interface())
		}
	}
	sql += ") VALUES("
	for i := 0; i < l; i++ {
		if i != 0 {
			sql += ", "
		}
		sql += "?"
	}
	sql += ") ON DUPLICATE KEY UPDATE "
	for i, v := range updates {
		if i != 0 {
			sql += ", "
		}
		sql += v + " = ?"
		args = append(args, val.FieldByName(v).Interface())
	}
	return
}

// create table
func (this User) SqlGeneratorCreate() (sqls []string) {
	sql := "CREATE TABLE users (\n"
	typ := reflect.TypeOf(this)
	val := reflect.ValueOf(this)
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).CanInterface() {
			t := typ.Field(i)
			sql += t.Name + " "
			switch t.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				sql += "INT, "
			case reflect.String:
				sql += "VARCHAR(255), "
			case reflect.Slice:
				sql += "BLOB, "
			default:
				panic("can not generate sql to create table")
			}
			sql += "\n"
		}
	}
	sql += "PRIMARY KEY (Uid)\n"
	sql += ") ENGINE=innoDB;\n"
	sql += "\n"
	sqls = make([]string, 0)
	sqls = append(sqls, sql)
	sqls = append(sqls, "CREATE UNIQUE INDEX uni_index_email ON users (Email);")
	sqls = append(sqls, "CREATE UNIQUE INDEX uni_index_nickname ON users (Nickname);")
	return
}
