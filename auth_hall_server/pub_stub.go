package main

import (
	"fmt"
	"math"
	"regexp"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

const (
	sessKeyUserId     = "userId"
	sessKeyRegister   = "register"
	sessKeyForgetPass = "forget"
	sessKeyEmail      = "email"
	minWithdraw       = 1
	minEnergy         = 1
	ratioEnergy2mBTC  = 10
	maxAvatar         = 1 << 18 // 256KB
	defaultEnergy     = 10

	// error const
	errUserNotExist    = "用户 %v 不存在"
	errTableNotExist   = "桌子 %v 不存在, 请加入别的桌子"
	errUnmatchedEmail  = "验证码发到邮箱 %s, 注册邮箱也必须是这个! 请不要这样攻击我们的服务!"
	errExceedMaxAvatar = "头像大小不能超过256KB, 该头像大小为 %dKB."
)

var (
	// regular expressions
	emailReg = regexp.MustCompile(`^[\w.]+\@[\w-]+\.[\w.]+$`)
	passReg  = regexp.MustCompile(`^[\w]{6,22}$`)
	nickReg  = regexp.MustCompile(`^(\p{Han}|\w){2,8}$`)

	// errors
	errEncryptPassword           = fmt.Errorf("密码哈希错误")
	errIncorrectPwd              = fmt.Errorf("密码错误")
	errIncorrectEmailFormat      = fmt.Errorf("邮箱格式错误")
	errIncorrectNicknameFormat   = fmt.Errorf("昵称格式错误, 必须由2~8个中文,英文,数字组成")
	errIncorrectPasswordFormat   = fmt.Errorf("密码格式错误, 必须由6~22个英文,数字组成")
	errEmailExist                = fmt.Errorf("邮箱已经被占用")
	errNicknameExist             = fmt.Errorf("昵称已经被占用")
	errNotLoggedIn               = fmt.Errorf("请先登陆")
	errBalNotSufficient          = fmt.Errorf("余额不足")
	errInvalidBtcAddr            = fmt.Errorf("无效的比特币地址")
	errExceedMinWithdraw         = fmt.Errorf("最低提现额度是 1mBTC")
	errExceedMinEnergy           = fmt.Errorf("最低购买1mBTC能量, 每1mBTC可以充%v能量", ratioEnergy2mBTC)
	errRegisterGetCodeFirst      = fmt.Errorf("需要先发送验证码到邮箱, 才能完成注册")
	errRegisterIncorrectCode     = fmt.Errorf("验证码错误, 请检查邮箱")
	errForgetPassGetCodeFirst    = fmt.Errorf("需要先发送验证码到邮箱, 才能找回密码")
	errForgetPassIncorrectCode   = fmt.Errorf("验证码错误, 请检查邮箱")
	errTableGameIsStarted        = fmt.Errorf("游戏已经开始, 请加入别的桌子")
	errTableIsFull               = fmt.Errorf("桌子已满, 请加入别的桌子进行游戏")
	errUnmatchNumOfFieldAndVal   = fmt.Errorf("两个数组长度不一样")
	errIncorrectType             = fmt.Errorf("更新字段类型只能是字符串, 整形, 二进制流[]byte")
	errAlreadyInGame             = fmt.Errorf("你已经在游戏中, 请先退出再加入另一个游戏")
	errNegativeBet               = fmt.Errorf("赌注不能为负数")
	errCantApplyForNilTournament = fmt.Errorf("暂无争霸赛, 无法加入.")
	errNilTournamentHall         = fmt.Errorf("暂无争霸赛, 无法获得争霸赛桌子信息")
	errCantMatchOpponent         = fmt.Errorf("无匹配对手, 请稍后重试.")
)

// send mail, register auth
func (pubStub) SendMailRegister(to string, ctx interface{}) error {
	authenCode := utils.RandString(8)
	session.SetSession(sessKeyRegister, authenCode, ctx)
	session.SetSession(sessKeyEmail, to, ctx)
	text := "请输入验证码: " + authenCode
	subject := "注册验证"
	pushFunc(func() {
		if err := utils.SendMail(text, subject, to); err != nil {
			log.Warn("send mail to %v error: %v", to, err)
			log.Warn("text is %v", text)
		}
	})
	return nil
}

// send mail, forget password
func (pubStub) SendMailForget(to string, ctx interface{}) error {
	if u := getUserByEmail(to); u == nil {
		return fmt.Errorf(errUserNotExist, to)
	}
	authenCode := utils.RandString(8)
	session.SetSession(sessKeyForgetPass, authenCode, ctx)
	session.SetSession(sessKeyEmail, to, ctx)
	text := "请输入验证码: " + authenCode
	subject := "找回密码"
	pushFunc(func() {
		if err := utils.SendMail(text, subject, to); err != nil {
			log.Warn("send mail to %v error: %v", to, err)
			log.Warn("text is %v", text)
		}
	})
	return nil
}

// forget password, modify it
func (pubStub) ForgetPassword(newPassword, authenCode string, ctx interface{}) error {
	// check authen code
	if code, ok := session.GetSession(sessKeyForgetPass, ctx).(string); !ok {
		return errForgetPassGetCodeFirst
	} else if code != authenCode {
		return errForgetPassIncorrectCode
	}

	if !passReg.MatchString(newPassword) {
		return errIncorrectPasswordFormat
	}
	// get user by email & update it
	email, ok := session.GetSession(sessKeyEmail, ctx).(string)
	if !ok {
		return errForgetPassGetCodeFirst
	}
	u := getUserByEmail(email)
	if u == nil {
		return fmt.Errorf(errUserNotExist, email)
	}
	if err := u.Update(types.NewUpdateString(types.UF_Password, utils.Encrypt(newPassword))); err != nil {
		return err
	}
	// delete session
	session.DeleteKey(sessKeyForgetPass, ctx)
	session.DeleteKey(sessKeyEmail, ctx)
	pushFunc(func() { insertOrUpdateUser(u) })
	return nil
}

// register
func (pubStub) Register(email, password, nickname, authenCode string, ctx interface{}) error {
	// check authen code
	if code, ok := session.GetSession(sessKeyRegister, ctx).(string); !ok {
		return errRegisterGetCodeFirst
	} else if code != authenCode {
		return errRegisterIncorrectCode
	}
	// check email
	if tmpEmail, ok := session.GetSession(sessKeyEmail, ctx).(string); !ok {
		return errRegisterGetCodeFirst
	} else if email != tmpEmail {
		return fmt.Errorf(errUnmatchedEmail, tmpEmail)
	}

	// check input
	if !emailReg.MatchString(email) {
		return errIncorrectEmailFormat
	}
	if !nickReg.MatchString(nickname) {
		return errIncorrectNicknameFormat
	}
	if !passReg.MatchString(password) {
		return errIncorrectPasswordFormat
	}
	if getUserByEmail(email) != nil {
		return errEmailExist
	}
	if getUserByNickname(nickname) != nil {
		return errNicknameExist
	}
	// generate new bitcoin address for the user
	addr, err := getNewAddress(nickname)
	if err != nil {
		return err
	}
	// encrypt password
	password = utils.Encrypt(password)
	if password == "" {
		return errEncryptPassword
	}
	// delete the authen code
	session.DeleteKey(sessKeyRegister, ctx)
	session.DeleteKey(sessKeyEmail, ctx)
	// add user in cache
	u := types.NewUser(users.GetNextId(), email, password, nickname, addr)
	u.Update(types.NewUpdateInt(types.UF_Energy, defaultEnergy)) // new user get 10 energy
	users.Add(u)
	// async insert into database
	pushFunc(func() { insertOrUpdateUser(u) })
	return nil
}

// login
func (pubStub) Login(nickname, password string, ctx interface{}) (err error) {
	u := getUserByNickname(nickname)
	if u == nil {
		return fmt.Errorf(errUserNotExist, nickname)
	}
	if u.Password == utils.Encrypt(password) {
		session.SetSession(sessKeyUserId, u.Uid, ctx)
		return nil
	}
	return errIncorrectPwd
}

// logout
func (pubStub) Logout(ctx interface{}) {
	// set user to free stat
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		users.SetFree(uid)
	}
	session.DelSession(ctx)
}

// get user info
// update session timestamp
func (pubStub) GetUserInfo(ctx interface{}) (*types.User, error) {
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			return nil, fmt.Errorf(errUserNotExist, uid)
		}
		session.SetSession(sessKeyUserId, u.Uid, ctx)
		return u, nil
	}
	return nil, errNotLoggedIn
}

// update user password
func (pubStub) UpdateUserPassword(currPass, newPass string, ctx interface{}) error {
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u.Password != utils.Encrypt(currPass) {
			return errIncorrectPwd
		}
		if !passReg.MatchString(newPass) {
			return errIncorrectPasswordFormat
		}
		if err := u.Update(types.NewUpdateString(types.UF_Password, utils.Encrypt(newPass))); err != nil {
			return err
		}
		pushFunc(func() { insertOrUpdateUser(u) })
		return nil
	}
	return errNotLoggedIn
}

// update user avatar
func (pubStub) UpdateUserAvatar(avatar []byte, ctx interface{}) error {
	if l := len(avatar); l > maxAvatar {
		return fmt.Errorf(errExceedMaxAvatar, l)
	}
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if err := u.Update(types.NewUpdate2dByte(types.UF_Avatar, avatar)); err != nil {
			return err
		}
		pushFunc(func() { insertOrUpdateUser(u) })
		return nil
	}
	return errNotLoggedIn
}

// get halls
func (pubStub) GetHalls(ctx interface{}) (map[string]interface{}, error) {
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			return nil, fmt.Errorf(errUserNotExist, uid)
		}
		return map[string]interface{}{
			"normal":     normalHall.Wrap(),
			"tournament": tournamentHall.Wrap(),
		}, nil
	}
	return nil, errNotLoggedIn
}

// get normal table
func (pubStub) GetNormalTable(tid int, ctx interface{}) (map[string]interface{}, error) {
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			return nil, fmt.Errorf(errUserNotExist, uid)
		}
		return normalHall.GetTableById(tid).WrapTable(), nil
	}
	return nil, errNotLoggedIn
}

// get tournament table
func (pubStub) GetTournamentTable(tid int, ctx interface{}) (map[string]interface{}, error) {
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			return nil, fmt.Errorf(errUserNotExist, uid)
		}
		if tournamentHall == nil {
			return nil, errNilTournamentHall
		}
		return tournamentHall.GetTableById(tid).WrapTable(), nil
	}
	return nil, errNotLoggedIn
}

// join a normal game, play or observe
// actually it is just get a token
func (pubStub) Join(tid int, isOb bool, ctx interface{}) (string, error) {
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			return "", fmt.Errorf(errUserNotExist, uid)
		}
		session.SetSession(sessKeyUserId, uid, ctx)
		if users.IsBusyUser(uid) {
			return "", errAlreadyInGame
		}
		t := normalHall.GetTableById(tid)
		if t == nil {
			return "", fmt.Errorf(errTableNotExist, tid)
		}
		if u.GetBalance() < t.GetBet() {
			return "", errBalNotSufficient
		}
		if u.GetEnergy() <= 0 {
			return "", errInsufficientEnergy
		}
		if !isOb {
			if t.IsStart() {
				return "", errTableGameIsStarted
			}
			if t.IsFull() {
				return "", errTableIsFull
			}
		}
		return utils.GenerateToken(uid, u.Nickname, false, isOb, tid)
	}
	return "", errNotLoggedIn
}

// TODO:
// observe a tournament game
// actually it is just get a token
func (pubStub) ObserveTournament(tid int, ctx interface{}) (string, error) {
	if tournamentHall == nil {
		return "", errNilTournamentHall
	}
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			return "", fmt.Errorf(errUserNotExist, uid)
		}
		session.SetSession(sessKeyUserId, uid, ctx)
		if users.IsBusyUser(uid) {
			return "", errAlreadyInGame
		}
		t := tournamentHall.GetTableById(tid)
		if t == nil {
			return "", fmt.Errorf(errTableNotExist, tid)
		}
		if t.IsStart() {
			return "", errTableGameIsStarted
		}
		return utils.GenerateToken(uid, u.Nickname, false, true, tid)
	}
	return "", errNotLoggedIn
}

// automatically match a normal game for user
func (pubStub) AutoMatch(ctx interface{}) (host string, token string, err error) {
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			err = fmt.Errorf(errUserNotExist, uid)
			return
		}
		session.SetSession(sessKeyUserId, uid, ctx)
		if users.IsBusyUser(uid) {
			err = errAlreadyInGame
			return
		}
		if u.GetEnergy() <= 0 {
			err = errInsufficientEnergy
			return
		}
		var table *types.Table = nil
		var gap float64 = 100
		for _, t := range normalHall.Tables.Tables {
			if t.IsFull() {
				continue
			}
			if u.GetBalance() < t.GetBet() {
				err = errBalNotSufficient
				return
			}
			if id := t.Get1pUid(); id != -1 {
				if tGap := math.Abs(float64(getUserById(id).Level - u.Level)); tGap < gap {
					gap = tGap
					table = t
				}
			}
		}
		if table == nil {
			err = errCantMatchOpponent
			return
		}
		token, err = utils.GenerateToken(uid, u.Nickname, false, false, table.TId)
		return table.GetHost(), token, err
	}
	err = errNotLoggedIn
	return
}

// create a game
func (pubStub) Create(title string, bet int, ctx interface{}) (int, error) {
	if bet < 0 {
		return -1, errNegativeBet
	}
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			return -1, fmt.Errorf(errUserNotExist, uid)
		}
		if users.IsBusyUser(uid) {
			return -1, errAlreadyInGame
		}
		if u.GetBalance() < bet {
			return -1, errBalNotSufficient
		}
		if u.GetEnergy() <= 0 {
			return -1, errInsufficientEnergy
		}
		id := normalHall.NextTableId()
		ip := clients.BestServer()
		host := ip + ":" + gameServerSocketPort
		if err := clients.GetStub(ip).Create(id, title, bet); err != nil {
			return -1, err
		}
		return id, normalHall.NewTable(id, title, host, bet)
	}
	return -1, errNotLoggedIn
}

// TODO:
// apply for a tournament
func (pubStub) Apply(ctx interface{}) (host, token string, err error) {
	if tournamentHall == nil {
		err = errCantApplyForNilTournament
		return
	}
	if uid, ok := session.GetSession(sessKeyUserId, ctx).(int); ok {
		u := getUserById(uid)
		if u == nil {
			err = fmt.Errorf(errUserNotExist, uid)
			return
		}
		if users.IsBusyUser(uid) {
			err = errAlreadyInGame
			return
		}
		session.SetSession(sessKeyUserId, uid, ctx)
		token, err = utils.GenerateToken(uid, u.Nickname, true, false, -1)
		return
	}
	err = errNotLoggedIn
	return
}

// withdraw
func (pubStub) Withdraw(amount int, address string, ctx interface{}) (string, error) {
	// check if amount >= minWithdraw
	if amount < minWithdraw {
		return "", errExceedMinWithdraw
	}
	// check btc address
	if isValid, err := validateAddress(address); err != nil {
		return "", err
	} else if !isValid {
		return "", errInvalidBtcAddr
	}
	// check if user logged in
	uid, ok := session.GetSession(sessKeyUserId, ctx).(int)
	if !ok {
		return "", errNotLoggedIn
	}

	u := getUserById(uid)
	if u == nil {
		return "", fmt.Errorf(errUserNotExist, uid)
	}
	// check balance
	if u.GetBalance() < amount {
		return "", errBalNotSufficient
	}
	// send coin
	txid, err := sendBitcoin(address, amount)
	if err != nil {
		return "", err
	}
	// update user in the cache
	if err := u.Update(types.NewUpdateInt("Balance", u.Balance-amount)); err != nil {
		return "", err
	}
	pushFunc(func() { insertWithdraw(txid, u.Nickname, address, amount) })
	return txid, nil
}

// buy energy
func (pubStub) BuyEnergy(amountOfmBTC int, ctx interface{}) error {
	// check if amountOfmBTC >= minEnergy
	if amountOfmBTC < minEnergy {
		return errExceedMinEnergy
	}
	// check if user logged in
	uid, ok := session.GetSession(sessKeyUserId, ctx).(int)
	if !ok {
		return errNotLoggedIn
	}
	session.SetSession(sessKeyUserId, uid, ctx)

	u := getUserById(uid)
	if u == nil {
		return fmt.Errorf(errUserNotExist, uid)
	}
	// check balance
	if u.GetBalance() < amountOfmBTC {
		return errBalNotSufficient
	}
	// update user in the cache
	if err := u.Update(types.NewUpdateInt(types.UF_Balance, u.GetBalance()-amountOfmBTC),
		types.NewUpdateInt(types.UF_Energy, u.GetEnergy()+amountOfmBTC*ratioEnergy2mBTC)); err != nil {
		return err
	}
	pushFunc(func() { buyEnergy(u.Uid, amountOfmBTC) })
	return nil
}
