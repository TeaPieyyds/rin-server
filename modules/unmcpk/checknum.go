package unmcpk

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
)

func md5Hex(parts ...string) string {
	h := md5.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func extractS1S2FromSource(code string) (string, string, error) {
	re := regexp.MustCompile(`message = '([^']*)' \+ data \+ '([^']*)'`)
	m := re.FindStringSubmatch(code)
	if len(m) != 3 {
		return "", "", fmt.Errorf("pattern not found")
	}
	return m[1], m[2], nil
}

func extractS1S2FromRepairedByRegex(repaired []byte) (string, string, error) {
	const startIndex = 251
	endIndex := 0
	for i := startIndex; i < len(repaired); i++ {
		if repaired[i] == 0x28 {
			endIndex = i
			break
		}
	}
	if endIndex <= startIndex {
		return "", "", fmt.Errorf("pattern not found")
	}

	target := repaired[startIndex:endIndex]
	re := regexp.MustCompile(`(?s)s.\x00{3}`)
	matches := re.FindIndex(target)
	if matches == nil {
		return "", "", fmt.Errorf("pattern not found")
	}
	s1 := string(target[:matches[0]])
	s2 := string(target[matches[1]:])
	return s1, s2, nil
}

func GenerateTransferCheckNum(isPC bool, data, engineVersion, patchVersion, python3Path string) (string, error) {
	var arr []any
	if err := json.Unmarshal([]byte(data), &arr); err != nil || len(arr) < 3 {
		return "", fmt.Errorf("bad data")
	}

	val, _ := arr[1].(string)
	uniqueFloat, _ := arr[2].(float64)
	uniqueID := int64(uniqueFloat)
	mcpHex, _ := arr[0].(string)
	mcpBytes, err := hex.DecodeString(mcpHex)
	if err != nil {
		return "", fmt.Errorf("bad mcp hex")
	}

	if python3Path == "" {
		python3Path = "python3"
	}

	const useRegex = true
	var s1, s2 string
	if useRegex {
		repaired, err := DecryptDynamicMCP(mcpBytes)
		if err != nil {
			return "", fmt.Errorf("decompile failed")
		}
		s1, s2, err = extractS1S2FromRepairedByRegex(repaired)
		if err != nil {
			return "", fmt.Errorf("pattern not found")
		}
	} else {
		source, err := DecompileDynamicMCP(mcpBytes, python3Path)
		if err != nil {
			return "", fmt.Errorf("decompile failed")
		}
		s1, s2, err = extractS1S2FromSource(source)
		if err != nil {
			return "", fmt.Errorf("pattern not found")
		}
	}

	valm := md5Hex(s1, val+"0", s2)
	tmps := make([]string, 0, len(valm)+7)
	for _, ch := range valm {
		tmps = append(tmps, fmt.Sprintf("%d", ((int(ch))*2+5)^255))
	}
	if !isPC {
		tmps = append(tmps, engineVersion, "android", patchVersion, "android", "2", "12", fmt.Sprintf("%d", uniqueID))
	} else {
		tmps = append(tmps, engineVersion, "windows", engineVersion, "win32", "0", "12", fmt.Sprintf("%d", uniqueID))
	}
	tmpsStr := ""
	for _, s := range tmps {
		tmpsStr += s
	}
	tmpsNum := md5Hex(s1, tmpsStr, s2)

	raw := valm[16:] + "False[]3" + tmpsNum + valm[:16]
	sign := md5Hex(s1, raw, s2)

	valueJSON := fmt.Sprintf("[\"%s\",\"%s\",false,[],\"\",\"\",3,\"%s\"]", valm, sign, tmpsNum)
	return valueJSON, nil
}
