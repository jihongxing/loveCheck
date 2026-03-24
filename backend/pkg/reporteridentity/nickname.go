package reporteridentity

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// Adjectives and nouns are combined deterministically from ReporterHash so the
// same reporter always gets the same virtual identity across reports.
var adjectives = []string{
	"匿名", "独行", "沉默", "坚毅", "深夜", "雨中", "理性", "执着",
	"清醒", "冷静", "敏锐", "低调", "审慎", "正直", "温和", "果敢",
	"朴素", "坦诚", "克制", "敏锐",
}

var nouns = []string{
	"守护者", "旅人", "记录者", "证人", "行者", "观察者", "之声", "微光",
	"哨兵", "执笔", "见证", "回声", "灯塔", "方舟", "信标", "锚点",
	"季风", "星辰", "潮汐", "原野",
}

// NicknameFromHash builds a stable display label from the 64-char hex HMAC.
func NicknameFromHash(hashHex string) string {
	if len(hashHex) < 16 {
		return "匿名用户"
	}
	raw := make([]byte, 8)
	if _, err := hex.Decode(raw, []byte(hashHex[:16])); err != nil {
		return "匿名用户"
	}
	u := binary.BigEndian.Uint64(raw)
	ai := int(u % uint64(len(adjectives)))
	ni := int((u / uint64(len(adjectives))) % uint64(len(nouns)))
	suffix := strings.ToUpper(hashHex[len(hashHex)-4:])
	return fmt.Sprintf("%s的%s #%s", adjectives[ai], nouns[ni], suffix)
}
