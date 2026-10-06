package g79client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// 在线大厅房间详情实体（部分字段）
type OnlineLobbyRoomGetEntity struct {
	RoomID     Uncertain `json:"room_id"`
	EntityID   Uncertain `json:"entity_id"`
	RoomName   string    `json:"room_name"`
	Slogan     string    `json:"slogan"`
	OwnerID    Uncertain `json:"owner_id"`
	ResID      Uncertain `json:"res_id"`
	CurNum     Uncertain `json:"cur_num"`
	MaxCount   Uncertain `json:"max_count"`
	Password   Uncertain `json:"password"`
	SaveID     string    `json:"save_id"`
	Version    string    `json:"version"`
	Visibility Uncertain `json:"visibility"`
}

type OnlineLobbyRoomGetResponse struct {
	Response
	Entity OnlineLobbyRoomGetEntity `json:"entity"`
}

// 获取在线大厅房间详情（按 room_id）
func (c *Client) GetOnlineLobbyRoom(roomID string) (*OnlineLobbyRoomGetResponse, error) {
	api := "/online-lobby-room/get"

	requestData := map[string]interface{}{
		"room_id": roomID,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.ReleaseJSON.ApiGatewayUrl+api, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "libhttpclient/1.0.0.0")
	req.Header.Set("user-id", c.UserID)

	token := CalculateDynamicToken(api, string(jsonData), c.UserToken)
	req.Header.Set("user-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := readResponseBody(resp)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DBG-GETROOM] body=%s\n", string(respBody))

	var getResp OnlineLobbyRoomGetResponse
	if err := json.Unmarshal(respBody, &getResp); err != nil {
		return nil, fmt.Errorf("解析在线大厅详情响应失败: %v, 响应内容: %s", err, string(respBody))
	}
	return &getResp, nil
}
