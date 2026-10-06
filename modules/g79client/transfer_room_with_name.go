package g79client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// TransferRoomWithName 请求
type TransferRoomWithNameRequest struct {
	Name string `json:"name"`
	UID  int64  `json:"uid"`
}

// TransferRoomWithNameListItem 简化的房间条目（可按需求扩展）
type TransferRoomWithNameListItem struct {
	HID          Uncertain   `json:"hid"`
	RID          Uncertain   `json:"rid"`
	Name         string      `json:"name"`
	Type         Uncertain   `json:"type"`
	Cnt          Uncertain   `json:"cnt"`
	TagIDs       []Uncertain `json:"tag_ids"`
	NeedPassword Uncertain   `json:"need_password"`
	Slogan       string      `json:"slogan"`
	Members      []Uncertain `json:"members"`
	RecommendV2  Uncertain   `json:"recommend_v2"`
	ModIDList    []Uncertain `json:"mod_id_list"`
	OwnerPing    Uncertain   `json:"owner_ping"`
	PerfLv       Uncertain   `json:"perf_lv"`
	RoomUniqueID string      `json:"room_unique_id"`
	Cap          Uncertain   `json:"cap"`
	SRV          Uncertain   `json:"srv"`
	Tips         string      `json:"tips"`
	Version      Uncertain   `json:"version"`
	Platform     Uncertain   `json:"platform"`
	ItemIDs      []Uncertain `json:"item_ids"`
	DeveloperID  Uncertain   `json:"developer_id"`
	Level        Uncertain   `json:"level"`
	PVP          Uncertain   `json:"pvp"`
	Permission   Uncertain   `json:"permission"`
	CreatedAt    Uncertain   `json:"created_at"`
}

type TransferRoomWithNameResponse struct {
	Response
	List []TransferRoomWithNameListItem `json:"list"`
}

// GetTransferRoomWithName 根据房间名/房间号查询（携带 user-token 动态签名）
func (c *Client) GetTransferRoomWithName(name string) (*TransferRoomWithNameResponse, error) {
	apiBase := c.ReleaseJSON.TransferServerNewHttpUrl
	api := "/room-with-name"
	userID, err := strconv.ParseInt(c.UserID, 10, 64)
	if err != nil {
		return nil, err
	}

	reqBody := TransferRoomWithNameRequest{
		Name: name,
		UID:  userID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiBase+api, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "libhttpclient/1.0.0.0")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", c.UserID)

	token := CalculateDynamicToken(api, string(jsonData), c.UserToken)
	req.Header.Set("user-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := readResponseBody(resp)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DBG-ROOM] url=%s body=%s\n", apiBase+api, string(body))

	var result TransferRoomWithNameResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 room-with-name 响应失败: %v, 响应内容: %s", err, string(body))
	}
	return &result, nil
}
