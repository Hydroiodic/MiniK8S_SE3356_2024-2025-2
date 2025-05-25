package ipalloc

import (
	"fmt"
	"strconv"
	"strings"
)

type IpAllocator struct {
	subnet   [4]int64
	mask_bit int64
	bitmap   []int64
}

func (ipa *IpAllocator) Init(sn string, mb string) error {
	// 解析子网掩码位数
	maskBits, err := strconv.ParseInt(mb, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid mask bits: %v", err)
	}

	if maskBits < 0 || maskBits > 32 {
		return fmt.Errorf("mask bits must be between 0 and 32")
	}

	// 初始化位图
	ipa.bitmap = make([]int64, 1<<(32-maskBits)) // 2^(32-maskBits)

	// 解析子网地址
	parts := strings.Split(sn, ".")

	if len(parts) != 4 {
		return fmt.Errorf("invalid subnet format, expected 'x.x.x.x'")
	}

	for i := 0; i < 4; i++ {
		ipa.subnet[i], err = strconv.ParseInt(parts[i], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid subnet octet %q: %v", parts[i], err)
		}

		if ipa.subnet[i] < 0 || ipa.subnet[i] > 255 {
			return fmt.Errorf("subnet octet %d must be 0-255", ipa.subnet[i])
		}
	}

	ipa.mask_bit = maskBits

	return nil
}

func (ipa *IpAllocator) AllocateIp() string {
	for i := int64(0); i < int64(len(ipa.bitmap)); i++ {
		if ipa.bitmap[i] == 0 {
			// 标记为已分配
			ipa.bitmap[i] = 1

			// 计算分配的IP地址
			// 先将子网部分转换为一个32位数（int64）
			subnetNum := ipa.subnet[0]<<24 | ipa.subnet[1]<<16 | ipa.subnet[2]<<8 | ipa.subnet[3]
			// 分配的IP = 子网部分 + i
			ipNum := subnetNum + i

			// 将ipNum拆回4个字节
			b0 := (ipNum >> 24) & 0xFF
			b1 := (ipNum >> 16) & 0xFF
			b2 := (ipNum >> 8) & 0xFF
			b3 := ipNum & 0xFF

			// 返回点分十进制格式
			return fmt.Sprintf("%d.%d.%d.%d", b0, b1, b2, b3)
		}
	}

	return ""
}

func (ipa *IpAllocator) DeallocateIp(ip string) error {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return fmt.Errorf("invalid IP format: %s", ip)
	}

	var ipOctets [4]int64

	for i := 0; i < 4; i++ {
		octet, err := strconv.ParseInt(parts[i], 10, 64)
		if err != nil || octet < 0 || octet > 255 {
			return fmt.Errorf("invalid IP octet %q in %s", parts[i], ip)
		}

		ipOctets[i] = octet
	}

	ipNum := ipOctets[0]<<24 | ipOctets[1]<<16 | ipOctets[2]<<8 | ipOctets[3]
	subnet := ipa.subnet[0]<<24 | ipa.subnet[1]<<16 | ipa.subnet[2]<<8 | ipa.subnet[3]
	index := ipNum - subnet

	if index < 0 || index >= int64(len(ipa.bitmap)) {
		return fmt.Errorf("invalid IP format: %s", ip)
	}

	if ipa.bitmap[index] == 0 {
		return fmt.Errorf("IP %s is not allocated", ip)
	}

	ipa.bitmap[index] = 0

	return nil
}
