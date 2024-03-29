package user

import (
	"devops-http/app/contract"
	"devops-http/app/module/base"
	"devops-http/app/module/base/request"
	"devops-http/app/module/base/response"
	"devops-http/app/module/base/utils"
	"devops-http/app/module/sys/model/config"
	"devops-http/app/module/sys/model/role"
	"devops-http/app/module/sys/model/user"
	"devops-http/framework"
	contract2 "devops-http/framework/contract"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	repository *base.Repository
}

func NewService(c framework.Container) *Service {
	db, err := c.MustMake(contract2.ORMKey).(contract2.ORMService).GetDB()
	logger := c.MustMake(contract2.LogKey).(contract2.Log)
	if err != nil {
		logger.Error("Service 获取db出错： err", zap.Error(err))
	}
	db.AutoMigrate(&base.DevopsSysUser{})
	return &Service{base.NewRepository(db)}
}

func (s *Service) GetRepository() *base.Repository {
	return s.repository
}

func (s *Service) GetUsers() {

}

func (s *Service) Login(req request.LoginRequest, jwt contract.JWTService) (interface{}, error) {
	result := make(map[string]string, 1)
	passwd, err := base64.StdEncoding.DecodeString(req.Password)
	password := string(passwd)
	var userData base.DevopsSysUser
	if req.Type == 0 || req.Type == 1 {
		// 其他类型暂时都是本地账户登录
		password = utils.MD5V([]byte(password))
		err = s.GetRepository().GetDB().Where("username = ? AND password = ?", req.Username, password).First(&userData).Error
		if err != nil {
			return result, errors.Errorf("用户名或者密码不正确：%v", err)
		}
	} else {
		// 域账户登录

	}
	// 创建jwt
	claims := jwt.CreateClaims(contract.BaseClaims{
		UUID:     userData.UUID,
		ID:       userData.ID,
		Username: userData.Username,
		NickName: userData.RealName,
	})
	token, err := jwt.CreateToken(claims)
	if err != nil {
		return result, err
	}
	result["token"] = token
	return result, err
}

func (s *Service) Modify(req user.DevopsSysUserEntity, c contract.Cabin) (interface{}, error) {
	var oldUser base.DevopsSysUser
	err := s.repository.SetRepository(&base.DevopsSysUser{}).Find(&oldUser, "id = ?", req.ID)
	if err != nil {
		return nil, errors.Errorf("未找到需要编辑的用户：%s", err.Error())
	}
	if req.Password != "" {
		passwd, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			return req.DevopsSysUser, err
		}
		req.DevopsSysUser.Password = utils.MD5V(passwd)
	}
	err = s.repository.SetRepository(&base.DevopsSysUser{}).Update(&req.DevopsSysUser, "id = ?", req.ID)
	req.Password = ""
	// 删除之前的角色
	_, err = c.GetCabin().DeleteRolesForUser(oldUser.UUID.String(), oldUser.Domain)
	if err != nil {
		err = errors.New("删除角色失败")
		return req.DevopsSysUser, err
	}
	// 添加角色
	_, err = c.GetCabin().AddRolesForUser(oldUser.UUID.String(), req.RoleIds, oldUser.Domain)
	if err != nil {
		err = errors.New("增加角色失败")
	}
	err = s.repository.SetRepository(&base.DevopsSysUser{}).Find(&oldUser, "id = ?", req.ID)
	return oldUser, err
}

func (s *Service) Add(req user.DevopsSysUserEntity, l contract.Ldap, c contract.Cabin) (interface{}, error) {
	userData := req.DevopsSysUser
	passwd, err := base64.StdEncoding.DecodeString(userData.Password)
	if err != nil {
		return userData, err
	}
	userData.Password = utils.MD5V(passwd)
	userData.UUID = uuid.NewV4()
	if !errors.Is(s.repository.GetDB().Where("name = ? ", req.Username).First(&config.DevopsSysConfig{}).Error, gorm.ErrRecordNotFound) {
		return userData, errors.New("存在相同用户名的用户")
	}
	if userData.UserType == 2 {
		filter := "OU=" + req.Domain
		if req.Domain != "freemud" {
			filter += ",OU=Merchants"
		}
		// ad 账户
		err = l.CreateUser(userData.Username, fmt.Sprintf("%v,DC=office,DC=freemud,DC=cn", filter), string(passwd), nil)
		if err != nil {
			return userData, errors.Errorf("AD 账户新增失败：%s", err.Error())
		}
	}
	err = s.repository.SetRepository(&base.DevopsSysUser{}).Save(&userData)
	if err != nil {
		return nil, errors.Errorf("新增失败：%s", err.Error())
	}
	userData.Password = ""
	// 添加角色
	flag, err := c.GetCabin().AddRolesForUser(userData.UUID.String(), req.RoleIds, userData.Domain)
	if !flag {
		err = errors.New("增加角色失败")
	}
	return userData, err
}

func (s *Service) Delete(ids string, c contract.Cabin) error {
	var users []base.DevopsSysUser
	err := s.repository.SetRepository(&base.DevopsSysUser{}).Find(&users, "id in (?)", ids)
	if err != nil {
		return errors.Errorf("未找到需要删除的用户：%s", err.Error())
	}
	if len(users) <= 0 {
		return errors.Errorf("未找到需要删除的用户")
	}
	for _, sysUser := range users {
		//if sysUser.UserType == 1 || sysUser.UserType == 0 {
		//
		//}
		err = s.repository.SetRepository(&base.DevopsSysUser{}).GetDB().Unscoped().Where("id = ?", sysUser.ID).Delete(&sysUser).Error
		if err != nil {
			return errors.Errorf("数据库删除用户: %s出错：%s", sysUser.Username, err.Error())
		}
		// 删除之前的角色
		_, err = c.GetCabin().DeleteRolesForUser(sysUser.UUID.String(), sysUser.Domain)
	}
	return err
}

func (s *Service) ChangePassword(req request.ChangePasswordRequest, l contract.Ldap) (err error) {
	if req.Type == 2 {
		_, err = l.Login(req.Username, req.OldPassword)
		if err != nil {
			return errors.Errorf("原密码不正确: %s", err)
		}
		err = l.ChangePassword(req.Username, req.Password)
		if err != nil {
			return errors.Errorf("密码修改失败: %s", err)
		}
	}
	return
}

// UserList 获取用户列表
func (s *Service) UserList(e contract.Cabin, req request.SearchUserParams) (response.PageResult, error) {
	list := make([]base.DevopsSysUser, 0)
	result := response.PageResult{
		List:     nil,
		Columns:  nil,
		Total:    0,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	db := s.repository.GetDB().Model(&base.DevopsSysUser{})
	if req.Username != "" {
		db.Where("username like ?", "%"+req.Username+"%")
	}

	err := db.Count(&result.Total).Error
	if err != nil {
		return result, err
	} else {
		db = db.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))
		if req.OrderKey != "" {
			var OrderStr string
			// 设置有效排序key 防止sql注入
			// 感谢 Tom4t0 提交漏洞信息
			orderMap := make(map[string]bool, 4)
			orderMap["id"] = true
			orderMap["username"] = true
			if orderMap[req.OrderKey] {
				if req.Desc {
					OrderStr = req.OrderKey + " desc"
				} else {
					OrderStr = req.OrderKey
				}
			} else { // didn't matched any order key in `orderMap`
				err = fmt.Errorf("非法的排序字段: %v", req.OrderKey)
				return result, err
			}
			err = db.Order(OrderStr).Find(&list).Error
		} else {
			err = db.Order("id").Find(&list).Error
		}
	}
	var viewList []user.DevopsSysUserView
	for i := range list {
		var viewUser user.DevopsSysUserView
		s.repository.GetDB().Model(&list[i]).Association("Groups").Find(&list[i].Groups)
		roleList, _ := e.GetCabin().GetRolesForUser(list[i].UUID.String(), list[i].Domain)
		viewUser.RoleIds = roleList
		viewUser.DevopsSysUser = list[i]
		viewUser.Password = ""
		// 查询角色名
		if len(roleList) > 0 {
			s.repository.GetDB().Model(&role.DevopsSysRole{}).Select("name").Find(&viewUser.Roles, roleList)
		}
		viewList = append(viewList, viewUser)
	}
	result.List = viewList
	result.Columns = user.SysUserViewColumns
	return result, err
}

// UserInfo 获取用户详细信息
func (s *Service) UserInfo(token *base.TokenUser, e contract.Cabin, filter []interface{}) (user.DevopsSysUserView, error) {
	res := base.DevopsSysUser{}
	err := s.repository.SetRepository(&base.DevopsSysUser{}).GetDB().First(&res, filter...).Error
	resView := user.DevopsSysUserView{DevopsSysUser: res}
	if res.ID <= 0 {
		return resView, errors.Errorf("未找到该用户： %v ！", err)
	}
	roleList, _ := e.GetCabin().GetRolesForUser(resView.UUID.String(), token.CurrentDomain)
	resView.RoleIds = roleList
	resView.Roles = make([]string, 0)
	var roleData []role.DevopsSysRole
	s.repository.SetRepository(&role.DevopsSysRole{}).Find(&roleData, roleList)
	for i := range roleData {
		resView.Roles = append(resView.Roles, roleData[i].Name)
	}
	return resView, err
}
