package drivenadapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/user_management.go -destination ../tests/mock_drivenadapters/user_management_mock.go

// UserManagement method interface
type UserManagement interface {
	// GetUserMailList 获取用户邮箱列表
	GetUserMailList(userID []string) ([]string, error)

	// 获取指定所有部门的邮箱
	GetDeptMailList(departmentIDs []string) ([]string, error)

	// GetGroupUserList 获取用户组下的所有用户
	GetGroupUserList(groupID []string) ([]string, error)

	// BatchGetUserInfo 批量获取用户名
	BatchGetUserInfo(userIDs []string) ([]UserInfo, error)

	// GetUserInfo 获取用户信息
	GetUserInfo(userID string) (UserInfo, error)

	// 根据accessorIDs获取names
	GetNameByAccessorIDs(accessorIDs map[string]string) (map[string]string, error)

	// 获取用户的accessorids
	GetUserAccessorIDs(userID string) (accessorIDs []string, err error)

	// RegisterInternalAccount 注册内部账号
	RegisterInternalAccount(name, password string) (string, error)

	// QueryInternalAccount 查询内部账号
	QueryInternalAccount(id string) (string, error)

	// CreateInternalGroup 创建内部组
	CreateInternalGroup() (id string, err error)

	// DeleteInternalGroup 删除内部组
	DeleteInternalGroup(ids []string) (err error)

	// UpdateInternalGroupMember 更新内部组成员
	UpdateInternalGroupMember(groupID string, userIDs []string) (err error)

	// GetInternalGroupMembers 获取内部组成员
	GetInternalGroupMembers(groupID string) (users []string, err error)

	// GetDepartmentInfo 获取部门信息
	GetDepartmentInfo(departmentID string) (*DepartInfo, error)

	// GetDepartments 获取子部门列表 /user-management/v1/departments?level=0
	GetDepartments(level int) (*[]DepInfo, error)

	// GetDepartmentMemberIDs 获取部门成员列表 /user-management/v1/departments/{department_id}/member_ids
	GetDepartmentMemberIDs(deptID string) (*DepartmentMembers, error)

	// BatchGetNames 批量获取用户名,支持user_ids、department_ids、contactor_ids、group_ids、app_ids
	BatchGetNames(data map[string][]string) (*NamesInfo, error)

	// GetAppAccountInfo 获取应用账户信息
	GetAppAccountInfo(appID string) (AppAccountInfo, error)

	// GetUserInfoByType 根据用户类型获取用户信息, 当前已支持类型 user, app
	GetUserInfoByType(accessorID, accessorType string) (UserInfo, error)

	// IsApp 判断是否为应用账号
	IsApp(appID string) (bool, error)
}

type userManagement struct {
	adminAddress string
	log          commonLog.Logger
	httpClient   HTTPClient
}

var (
	uOnce sync.Once
	u     UserManagement
)

// UserInfo 用户信息
type UserInfo struct {
	UserID          string                 `json:"userid"`
	UserName        string                 `json:"username"`
	ParentDeps      interface{}            `json:"parentDeps"`
	ParentDepIDs    []string               `json:"parentDepIDs"`
	CsfLevel        float64                `json:"csfLevel"`
	UdID            string                 `json:"udid"`
	LoginIP         string                 `json:"ip"`
	TokenID         string                 `json:"-"`
	Roles           []string               `json:"roles"`
	Telephone       string                 `json:"telephone"`
	Email           string                 `json:"email"`
	Enabled         bool                   `json:"enabled"`
	Type            string                 `json:"type"`
	VisitorType     string                 `json:"visitor_type"` // 访问类型，已存在app、authenticated_user、anonymous_user
	UserAgent       string                 `json:"-"`
	CustomAttr      map[string]interface{} `json:"custom_attr"`
	IsKnowledge     bool                   `json:"is_knowledge"`
	DepartmentPaths []DepartmentPath       `json:"-"`
	DepartmentNames []string               `json:"-"`
	ClientType      string                 `json:"-"`
	AccountType     string                 `json:"-"` // 用户账户类型，目前已存在user、app
	Mac             string                 `json:"-"`
	ExpiresIn       int64                  `json:"-"` // 用于处理应用账户Token有效期
}

type DepartmentPath struct {
	ID   string `json:"id_path"`
	Name string `json:"name_path"`
}

// UserAttribute 用户属性
type UserAttribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NamesInfo struct {
	UserNames       []*UserAttribute `json:"user_names"`
	GroupNames      []*UserAttribute `json:"group_names"`
	DepartmentNames []*UserAttribute `json:"department_names"`
	ContactorNames  []*UserAttribute `json:"contactor_names"`
	AppNames        []*UserAttribute `json:"app_names"`
}

func (u *NamesInfo) ToMap(userType ...string) map[string]string {
	var res = map[string]string{}
	for _, val := range userType {
		switch val {
		case common.User.ToString():
			for _, val := range u.UserNames {
				res[val.ID] = val.Name
			}
		case common.Group.ToString():
			for _, val := range u.GroupNames {
				res[val.ID] = val.Name
			}
		case common.Department.ToString():
			for _, val := range u.DepartmentNames {
				res[val.ID] = val.Name
			}
		case common.Contactor.ToString():
			for _, val := range u.ContactorNames {
				res[val.ID] = val.Name
			}
		case common.APP.ToString():
			for _, val := range u.AppNames {
				res[val.ID] = val.Name
			}
		}
	}

	return res
}

func (u *UserInfo) SetFullDepPath(spliceType ...string) {
	if v, ok := u.ParentDeps.([]interface{}); ok {
		var names, ids []string
		for _, parentDep := range v {
			_parentDeps, ok := parentDep.([]interface{})
			if !ok {
				continue
			}
			for _, parentDep := range _parentDeps {
				dept_map, ok := parentDep.(map[string]interface{})
				if !ok {
					continue
				}
				names = append(names, dept_map["name"].(string))
				ids = append(ids, dept_map["id"].(string))
			}
			u.DepartmentPaths = append(u.DepartmentPaths, DepartmentPath{
				ID:   strings.Join(ids, "/"),
				Name: strings.Join(names, "/"),
			})
			u.DepartmentNames = append(u.DepartmentNames, names[len(names)-1])
			names = []string{}
			ids = []string{}
		}
	}
}

// DepartInfo 部门/组织信息
type DepartInfo struct {
	DepartmentID string    `json:"department_id"`
	Name         string    `json:"name"`
	ParentDeps   []DepInfo `json:"parent_deps"`
}

// DepInfo 组织信息
type DepInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// DepartmentMembers 部门成员
type DepartmentMembers struct {
	UserIDs       []string `json:"user_ids"`
	DepartmentIDs []string `json:"department_ids"`
}

type orgNameIDInfo struct {
	UserIDs      map[string]string
	DepartIDs    map[string]string
	ContactorIDs map[string]string
	GroupIDs     map[string]string
}

type orgIDInfo struct {
	UserIDs      []string
	DepartIDs    []string
	ContactorIDs []string
	GroupIDs     []string
}

// AppAccountInfo 应用账户信息
type AppAccountInfo struct {
	AppID string `json:"id"`
	Name  string `json:"name"`
}

// NewUserManagement 创建获取用户服务
func NewUserManagement() UserManagement {
	uOnce.Do(func() {
		config := common.NewConfig()
		if !config.IsAuthEnabled() {
			u = &mockUserManagement{}
		} else {
			u = &userManagement{
				adminAddress: fmt.Sprintf("http://%s:%v", config.UserManagement.PrivateHost, config.UserManagement.PrivatePort),
				log:          commonLog.NewLogger(),
				httpClient:   NewHTTPClient(),
			}
		}
	})
	return u
}

func (u *userManagement) GetGroupUserList(groupID []string) ([]string, error) {
	var users []string
	target := fmt.Sprintf("%s/api/user-management/v1/group-members", u.adminAddress)
	paras := map[string]interface{}{"method": "GET", "group_ids": groupID}
	_, respParam, err := u.httpClient.Post(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, paras)
	if err != nil {
		u.log.Errorf("GetGroupUserList failed: %v, url: %v", err, target)
		return nil, err
	}

	userIDs := respParam.(map[string]interface{})["user_ids"].([]interface{})
	for _, userID := range userIDs {
		users = append(users, userID.(string))
	}
	return users, nil
}

// 获取指定userid所有用户的邮箱
func (u *userManagement) GetUserMailList(userID []string) ([]string, error) {
	userIDs := strings.Join(userID, "&user_id=")
	target := fmt.Sprintf("%s/api/user-management/v1/emails?user_id=%s", u.adminAddress, userIDs)
	respParam, err := u.httpClient.Get(target, nil)
	if err != nil {
		u.log.Errorf("GetUserMailList failed: %v, url: %v", err, target)
		return nil, err
	}

	userMails := respParam.(map[string]interface{})["user_emails"].([]interface{})

	mails := make([]string, 0)
	for _, userMail := range userMails {
		userMail := userMail.(map[string]interface{})
		if userMail["email"].(string) != "" {
			mails = append(mails, userMail["email"].(string))
		}
	}
	return mails, nil
}

// 获取指定所有部门的邮箱
func (u *userManagement) GetDeptMailList(departmentIDs []string) ([]string, error) {
	departmentIDstr := strings.Join(departmentIDs, "&department_id=")
	target := fmt.Sprintf("%s/api/user-management/v1/emails?department_id=%s", u.adminAddress, departmentIDstr)
	respParam, err := u.httpClient.Get(target, nil)
	if err != nil {
		u.log.Errorf("GetDeptMailList failed: %v, url: %v", err, target)
		return nil, err
	}

	deptMails := respParam.(map[string]interface{})["department_emails"].([]interface{})

	mails := make([]string, 0)
	for _, deptMail := range deptMails {
		deptMail := deptMail.(map[string]interface{})
		if deptMail["email"].(string) != "" {
			mails = append(mails, deptMail["email"].(string))
		}
	}
	return mails, nil
}

func (u *userManagement) BatchGetUserInfo(userIDs []string) (users []UserInfo, err error) {
	target := fmt.Sprintf("%s/api/user-management/v1/names", u.adminAddress)
	body := map[string]interface{}{"method": "GET", "user_ids": userIDs}
	_, respParam, err := u.httpClient.Post(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		u.log.Errorf("BatchGetUserInfo failed: %v, url: %v", err, target)
		return nil, err
	}

	results := respParam.(map[string]interface{})["user_names"].([]interface{})
	for _, item := range results {
		userid := item.(map[string]interface{})["id"].(string)
		username := item.(map[string]interface{})["name"].(string)
		users = append(users, UserInfo{
			UserID:   userid,
			UserName: username,
		})
	}
	return
}

func (u *userManagement) GetUserInfo(userID string) (userInfo UserInfo, err error) {
	target := fmt.Sprintf("%s/api/user-management/v1/users/%s/name,parent_deps,csf_level,roles,email,telephone,enabled,custom_attr", u.adminAddress, userID)
	respParam, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		u.log.Errorf("GetUserInfo failed: %v, url: %v", err, target)
		return userInfo, err
	}

	userInfos, ok := respParam.([]interface{})

	if !ok {
		return
	}

	curUserInfo := userInfos[0]

	userInfo.UserID = userID
	userInfo.UserName = curUserInfo.(map[string]interface{})["name"].(string)
	userInfo.CsfLevel = curUserInfo.(map[string]interface{})["csf_level"].(float64)
	userInfo.Telephone = curUserInfo.(map[string]interface{})["telephone"].(string)
	userInfo.Email = curUserInfo.(map[string]interface{})["email"].(string)
	userInfo.Enabled = curUserInfo.(map[string]interface{})["enabled"].(bool)

	parentDeps := curUserInfo.(map[string]interface{})["parent_deps"].([]interface{})
	for _, dep := range parentDeps {
		r := dep.([]interface{})
		if len(r) > 0 {
			latestItem := r[len(r)-1]
			parsedItem := latestItem.(map[string]interface{})
			userInfo.ParentDepIDs = append(userInfo.ParentDepIDs, parsedItem["id"].(string))
		}
	}
	roles := curUserInfo.(map[string]interface{})["roles"].([]interface{})
	for _, v := range roles {
		r := v.(string)
		userInfo.Roles = append(userInfo.Roles, r)
	}
	userInfo.ParentDeps = parentDeps
	if customAttr, ok := curUserInfo.(map[string]interface{})["custom_attr"]; ok {
		if attrs, ok := customAttr.(map[string]interface{}); ok {
			userInfo.CustomAttr = attrs
			isKnowledge, ok := attrs["is_knowledge"]
			if ok {
				userInfo.IsKnowledge = fmt.Sprintf("%v", isKnowledge) == "1"
			}
		}
	}
	userInfo.SetFullDepPath()
	return
}

func (u *userManagement) GetUserAccessorIDs(userID string) (accessorIDs []string, err error) {
	target := fmt.Sprintf("%s/api/user-management/v1/users/%s/accessor_ids", u.adminAddress, userID)
	respParam, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		u.log.Errorf("GetUserAccessorIDs failed: %v, url: %v", err, target)
		return accessorIDs, err
	}

	ids, _ := respParam.([]interface{})
	for _, id := range ids {
		accessorIDs = append(accessorIDs, id.(string))
	}

	return accessorIDs, nil
}

func (u *userManagement) getOrgNameIDInfo(orgInfo *orgIDInfo) (orgNameInfo orgNameIDInfo, err error) {

	userIDs := orgInfo.UserIDs
	departIDs := orgInfo.DepartIDs
	contactorIDs := orgInfo.ContactorIDs
	groupIDs := orgInfo.GroupIDs

	target := fmt.Sprintf("%s/api/user-management/v1/names", u.adminAddress)

	var respParam interface{}

	for {
		tmpInfo := map[string]interface{}{
			"method":         "GET",
			"user_ids":       userIDs,
			"department_ids": departIDs,
			"contactor_ids":  contactorIDs,
			"group_ids":      groupIDs,
		}

		_, respParam, err = u.httpClient.Post(target, nil, tmpInfo)

		if err == nil {
			break
		}

		if httpErr, ok := err.(errors.ExHTTPError); ok {
			var errBody = make(map[string]interface{}, 0)
			jsonErr := json.Unmarshal([]byte(httpErr.Body), &errBody)

			if jsonErr != nil {
				u.log.Errorf("getOrgNameIDInfo failed: %v, url: %v", err, target)
				return
			}

			if errCode, ok := errBody["code"]; ok {
				if code, ok := errCode.(float64); ok && (code == 400019001 || code == 400019002 || code == 400019003 || code == 400019004) {
					for _, id := range errBody["detail"].(map[string]interface{})["ids"].([]interface{}) {
						userIDs = utils.StringExclude(userIDs, id.(string))
						departIDs = utils.StringExclude(departIDs, id.(string))
						contactorIDs = utils.StringExclude(contactorIDs, id.(string))
						groupIDs = utils.StringExclude(groupIDs, id.(string))
					}
					continue
				}
			}
		}

		u.log.Errorf("getOrgNameIDInfo failed: %v, url: %v", err, target)
		return
	}

	userNameInfos := respParam.(map[string]interface{})["user_names"].([]interface{})
	orgNameInfo.UserIDs = make(map[string]string)
	for _, x := range userNameInfos {
		id := x.(map[string]interface{})["id"].(string)
		name := x.(map[string]interface{})["name"].(string)
		orgNameInfo.UserIDs[id] = name
	}
	orgNameInfo.DepartIDs = make(map[string]string)
	departNameInfos := respParam.(map[string]interface{})["department_names"].([]interface{})
	for _, x := range departNameInfos {
		id := x.(map[string]interface{})["id"].(string)
		name := x.(map[string]interface{})["name"].(string)
		orgNameInfo.DepartIDs[id] = name
	}
	orgNameInfo.ContactorIDs = make(map[string]string)
	conatctorNameInfos := respParam.(map[string]interface{})["contactor_names"].([]interface{})
	for _, x := range conatctorNameInfos {
		id := x.(map[string]interface{})["id"].(string)
		name := x.(map[string]interface{})["name"].(string)
		orgNameInfo.ContactorIDs[id] = name
	}
	orgNameInfo.GroupIDs = make(map[string]string)
	groupNameInfos := respParam.(map[string]interface{})["group_names"].([]interface{})
	for _, x := range groupNameInfos {
		id := x.(map[string]interface{})["id"].(string)
		name := x.(map[string]interface{})["name"].(string)
		orgNameInfo.GroupIDs[id] = name
	}
	return
}

func (u *userManagement) GetNameByAccessorIDs(accessorIDs map[string]string) (accessorNames map[string]string, err error) { //nolint
	var orgInfo orgIDInfo
	orgInfo.UserIDs = make([]string, 0)
	orgInfo.DepartIDs = make([]string, 0)
	orgInfo.ContactorIDs = make([]string, 0)
	orgInfo.GroupIDs = make([]string, 0)
	for accessorID, accessorType := range accessorIDs {
		if accessorType == common.User.ToString() {
			orgInfo.UserIDs = append(orgInfo.UserIDs, accessorID)
		} else if accessorType == common.Department.ToString() {
			orgInfo.DepartIDs = append(orgInfo.DepartIDs, accessorID)
		} else if accessorType == common.Contactor.ToString() {
			orgInfo.ContactorIDs = append(orgInfo.ContactorIDs, accessorID)
		} else if accessorType == common.Group.ToString() {
			orgInfo.GroupIDs = append(orgInfo.GroupIDs, accessorID)
		}
	}

	orgNameInfo, err := u.getOrgNameIDInfo(&orgInfo)
	if err != nil {
		u.log.Errorf("GetNameByAccessorID err:%v", err)
		return
	}
	accessorNames = make(map[string]string)
	for accessorID, accessorType := range accessorIDs {
		if accessorType == common.User.ToString() {
			if value, ok := orgNameInfo.UserIDs[accessorID]; ok {
				accessorNames[accessorID] = value
			}
		} else if accessorType == common.Department.ToString() {
			if value, ok := orgNameInfo.DepartIDs[accessorID]; ok {
				accessorNames[accessorID] = value
			}
		} else if accessorType == common.Contactor.ToString() {
			if value, ok := orgNameInfo.ContactorIDs[accessorID]; ok {
				accessorNames[accessorID] = value
			}
		} else if accessorType == common.Group.ToString() {
			if value, ok := orgNameInfo.GroupIDs[accessorID]; ok {
				accessorNames[accessorID] = value
			}
		}
	}
	return
}

func (u *userManagement) RegisterInternalAccount(name, password string) (string, error) {
	target := fmt.Sprintf("%s/api/user-management/v1/apps", u.adminAddress)
	body := map[string]string{
		"name":     name,
		"type":     "internal",
		"password": password,
	}
	respCode, respParam, err := u.httpClient.Post(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		if respCode == http.StatusConflict {
			parsedError, err := ExHTTPErrorParser(err) //nolint
			if err != nil {
				u.log.Errorf("RegisterInternalAccount failed: %v, url: %v", err, target)
				return "", err
			}
			if detail, ok := parsedError["detail"].(map[string]interface{}); ok {
				if id, ok := detail["id"].(string); ok {
					return id, nil
				}
			}
			return "", err
		}
		u.log.Errorf("RegisterInternalAccount failed: %v, url: %v", err, target)
		return "", err
	}

	result := respParam.(map[string]interface{})["id"]
	id := result.(string)
	return id, nil
}

func (u *userManagement) QueryInternalAccount(id string) (string, error) {
	target := fmt.Sprintf("%s/api/user-management/v1/apps/%s", u.adminAddress, id)
	respParam, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		httpError, ok := err.(errors.ExHTTPError)
		if ok && httpError.Status == http.StatusNotFound {
			return "", nil
		}

		u.log.Errorf("QueryInternalAccount failed: %v, url: %v", err, target)
		return "", err
	}

	result := respParam.(map[string]interface{})["name"]
	name := result.(string)
	return name, nil
}

// CreateInternalGroup 创建内部组
func (u *userManagement) CreateInternalGroup() (id string, err error) {
	target := fmt.Sprintf("%s/api/user-management/v1/internal-groups", u.adminAddress)
	_, respParam, err := u.httpClient.Post(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, map[string]string{})
	if err != nil {
		u.log.Errorf("CreateInternalGroup failed: %v, url: %v", err, target)
		return
	}

	if res, ok := respParam.(map[string]interface{}); ok {
		id = res["id"].(string)
	}

	return
}

// DeleteInternalGroup 删除内部组
func (u *userManagement) DeleteInternalGroup(ids []string) (err error) {
	target := fmt.Sprintf("%s/api/user-management/v1/internal-groups/%s", u.adminAddress, strings.Join(ids, ","))
	_, err = u.httpClient.Delete(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		u.log.Errorf("CreateInternalGroup failed: %v, url: %v", err, target)
	}

	return
}

// UpdateInternalGroupMember 更新内部组成员
func (u *userManagement) UpdateInternalGroupMember(groupID string, userIDs []string) (err error) {
	target := fmt.Sprintf("%s/api/user-management/v1/internal-group-members/%s", u.adminAddress, groupID)
	body := []map[string]string{}
	for _, id := range userIDs {
		body = append(body, map[string]string{"id": id, "type": common.User.ToString()})
	}
	_, _, err = u.httpClient.Put(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		u.log.Errorf("UpdateInternalGroupMenmber failed: %v, url: %v", err, target)
	}

	return
}

// GetInternalGroupMembers 更新内部组成员
func (u *userManagement) GetInternalGroupMembers(groupID string) (users []string, err error) {
	target := fmt.Sprintf("%s/api/user-management/v1/internal-group-members/%s", u.adminAddress, groupID)
	respParam, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		u.log.Errorf("UpdateInternalGroupMenmber failed: %v, url: %v", err, target)
	}

	res, ok := respParam.([]interface{})
	if !ok {
		return
	}
	for index := range res {
		if user, isuser := res[index].(map[string]interface{}); isuser {
			users = append(users, user["id"].(string))
		}
	}

	return
}

// GetDepartmentInfo 获取部门信息 /user-management/v1/departments/{department_ids}/{fields}
func (u *userManagement) GetDepartmentInfo(departmentID string) (*DepartInfo, error) {
	target := fmt.Sprintf("%s/api/user-management/v1/departments/%s/name,parent_deps", u.adminAddress, departmentID)
	respParam, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		u.log.Errorf("GetDepartmentInfo failed: %v, url: %v", err, target)
		return nil, err
	}

	bytes, err := json.Marshal(respParam)
	if err != nil {
		u.log.Errorf("GetDepartmentInfo parsed info failed: %v, url: %v", err, target)
		return nil, err
	}

	var departInfos []DepartInfo

	err = json.Unmarshal(bytes, &departInfos)
	if err != nil {
		u.log.Errorf("GetDepartmentInfo parsed info failed: %v, url: %v", err, target)
		return nil, err
	}

	departInfo := departInfos[0]

	return &departInfo, nil
}

// GetDepartments 获取子部门列表 /user-management/v1/departments?level=0
func (u *userManagement) GetDepartments(level int) (*[]DepInfo, error) {
	target := fmt.Sprintf("%s/api/user-management/v1/departments?level=%d", u.adminAddress, level)
	respParam, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		u.log.Errorf("GetDepartments failed: %v, url: %v", err, target)
		return nil, err
	}

	bytes, err := json.Marshal(respParam)
	if err != nil {
		u.log.Errorf("GetDepartments parsed info failed: %v, url: %v", err, target)
		return nil, err
	}

	var depInfos []DepInfo

	err = json.Unmarshal(bytes, &depInfos)
	if err != nil {
		u.log.Errorf("GetDepartments parsed info failed: %v, url: %v", err, target)
		return nil, err
	}

	return &depInfos, nil
}

// GetDepartmentMemberIDs 获取部门成员列表 /user-management/v1/departments/{department_id}/member_ids
func (u *userManagement) GetDepartmentMemberIDs(deptID string) (*DepartmentMembers, error) {
	target := fmt.Sprintf("%s/api/user-management/v1/departments/%s/member_ids", u.adminAddress, deptID)
	respParam, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		u.log.Errorf("GetDepartmentMemberIDs failed: %v, url: %v", err, target)
		return nil, err
	}

	bytes, err := json.Marshal(respParam)
	if err != nil {
		u.log.Errorf("GetDepartmentMemberIDs parsed info failed: %v, url: %v", err, target)
		return nil, err
	}

	var deptMems DepartmentMembers

	err = json.Unmarshal(bytes, &deptMems)
	if err != nil {
		u.log.Errorf("GetDepartmentMemberIDs parsed info failed: %v, url: %v", err, target)
		return nil, err
	}

	return &deptMems, nil
}

// BatchGetNames 批量获取用户名
func (u *userManagement) BatchGetNames(data map[string][]string) (namesInfo *NamesInfo, err error) {
	target := fmt.Sprintf("%s/api/user-management/v1/names", u.adminAddress)
	body := map[string]interface{}{"method": "GET"}
	for key, value := range data {
		body[key] = value
	}
	_, respParam, err := u.httpClient.Post(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		u.log.Errorf("BatchGetUserInfo failed: %v, url: %v", err, target)
		return
	}

	bodyBytes, _ := json.Marshal(respParam)
	namesInfo = &NamesInfo{}
	err = json.Unmarshal(bodyBytes, &namesInfo)
	if err != nil {
		u.log.Errorf("BatchGetNames parsed info failed: %v, url: %v", err, target)
		return
	}

	return
}

// GetAppAccountInfo 获取应用账户信息
func (u *userManagement) GetAppAccountInfo(appID string) (AppAccountInfo, error) {
	var appInfo AppAccountInfo
	target := fmt.Sprintf("%s/api/user-management/v1/apps/%s", u.adminAddress, appID)
	respParam, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		u.log.Errorf("GetAppInfo failed: %v, url: %v", err, target)
		return appInfo, err
	}

	bytes, err := json.Marshal(respParam)
	if err != nil {
		u.log.Errorf("GetAppInfo parsed info failed: %v, url: %v", err, target)
		return appInfo, err
	}
	err = json.Unmarshal(bytes, &appInfo)
	if err != nil {
		u.log.Errorf("GetAppInfo parsed info failed: %v, url: %v", err, target)
		return appInfo, err
	}

	return appInfo, nil
}

// GetUserInfoByType 根据用户类型获取用户信息, 当前已支持类型 user, app
func (m *userManagement) GetUserInfoByType(accessorID, accessorType string) (UserInfo, error) {
	var userInfo UserInfo
	if accessorType == common.APP.ToString() {
		app, err := m.GetAppAccountInfo(accessorID)
		if err != nil {
			return userInfo, err
		}
		userInfo.UserID = app.AppID
		userInfo.UserName = app.Name
		userInfo.AccountType = common.APP.ToString()
		return userInfo, nil
	}

	return m.GetUserInfo(accessorID)
}

// IsApp 判断是否是应用账号
func (u *userManagement) IsApp(appID string) (bool, error) {
	target := fmt.Sprintf("%s/api/user-management/v1/apps/%s", u.adminAddress, appID)
	_, err := u.httpClient.Get(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		httpError, ok := err.(errors.ExHTTPError)
		if ok && httpError.Status == http.StatusNotFound {
			return false, nil
		}
		u.log.Errorf("Check is app failed: %v, url: %v", err, target)
		return false, err
	}

	return true, nil
}
