package g79client

import (
	"net/http"
	"strconv"
	"time"
)

var EngineVersion = "3.6.5.281774"

func Refetch() {
	packList, _ := RefreshG79PackList()
	// 版本已锁定，不覆盖
	_ = packList
	_, _ = RefreshG79LatestVersion()
	_, _ = RefreshG79ReleaseJSON()
	_, _ = RefreshX19ReleaseJSON()
	_, _ = RefreshG79ChatServers()
	_, _ = RefreshG79LinkServers()
	_, _ = RefreshG79TransferServers()
}

func init() {
	Refetch()
	go func() {
		for {
			time.Sleep(time.Second * 60)
			Refetch()
		}
	}()
}

// Client 结构体
type Client struct {
	UserID             string
	UserToken          string
	Seed               string
	ReleaseJSON        G79ReleaseJSON
	X19ReleaseJSON     X19ReleaseJSON
	EngineVersion      string
	G79LatestVersion   string
	patchResourcesHash string
	UserDetail         *UserDetailEntity
	peUserLoginAfter   *PeUserLoginAfterResponse
	httpClient         *http.Client
	Cookie             string
}

// 创建新的客户端
func NewClient() (*Client, error) {
	c := &Client{
		EngineVersion: EngineVersion,
		httpClient:    &http.Client{},
	}
	var err error

	// 锁死为注释中已验证配对的旧版本参数（3.6 系列）
	// 动态拉取最新 patch 版本与资源 hash（不再锁死旧版本）
	if pm, perr := GetGlobalG79PatchMetadata(); perr == nil && pm != nil {
		c.G79LatestVersion = pm.Version
		c.patchResourcesHash = pm.ResourcesHash
	}
	if c.G79LatestVersion == "" {
		c.G79LatestVersion = "3.6.56.288487"
	}
	if c.patchResourcesHash == "" {
		c.patchResourcesHash = "548ddfa2a67139a3e1b49dd812267445"
	}
	ReleaseJSON, err := GetGlobalG79ReleaseJSON()
	if err != nil {
		return nil, err
	}
	c.ReleaseJSON = *ReleaseJSON
	X19ReleaseJSON, err := GetGlobalX19ReleaseJSON()
	if err != nil {
		return nil, err
	}
	c.X19ReleaseJSON = *X19ReleaseJSON
	return c, nil
}

// 使用自定义 HTTP Client 创建客户端（支持代理）
func NewClientWithHTTPClient(hc *http.Client) (*Client, error) {
	c, err := NewClient()
	if err != nil {
		return nil, err
	}
	if hc != nil {
		c.httpClient = hc
	}
	return c, nil
}

// 设置用户凭证
func (c *Client) SetCredentials(userID, userToken string) {
	c.UserID = userID
	c.UserToken = userToken
}

// 获取用户ID的整数形式
func (c *Client) GetUserIDInt() (int64, error) {
	return strconv.ParseInt(c.UserID, 10, 64)
}
