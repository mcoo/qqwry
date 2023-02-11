/*
Copyright © 2023 enjoy <i@mcenjoy.cn>
*/
package dat

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Dat struct {
	fs *os.File
}

var loc = time.FixedZone("CST", 3600*8)

func New(fs *os.File) Dat {
	return Dat{fs: fs}
}

func (d Dat) SearchIp(ipv4 net.IP) (country, area string, e error) {
	index, err := d.searchIndex(IpToUint32(ipv4))
	if err != nil {
		return "", "", err
	}
	offset, err := d.readUint32WithLength3(int64(index) + 4)
	if err != nil {
		return "", "", err
	}
	return d.readAddress(int64(offset))
}
func (d Dat) Version() (*time.Time, error) {
	versionOffset, err := d.readUint32(4)
	if err != nil {
		return nil, err
	}
	versionOffsetOffset, err := d.readUint32WithLength3(int64(versionOffset) + 4)
	if err != nil {
		return nil, err
	}
	_, version, err := d.readAddress(int64(versionOffsetOffset))
	if err != nil {
		return nil, err
	}

	t, err := time.ParseInLocation("2006年01月02日IP数据", version, loc)
	if err != nil {
		t, err = time.ParseInLocation("2006年1月2日IP数据", version, loc)
		if err != nil {
			return nil, err
		}
	}
	return &t, nil
}
func (d Dat) readByte(offset int64) (byte, error) {
	var data byte
	err := d.readWithSize(offset, &data, 1, binary.LittleEndian)
	if err != nil {
		return 0, err
	}
	return data, nil
}
func (d Dat) readUint32(offset int64) (uint32, error) {
	var data uint32
	err := d.readWithSize(offset, &data, 4, binary.LittleEndian)
	if err != nil {
		return 0, err
	}
	return data, nil
}
func (d Dat) readUint32WithLength3(offset int64) (uint32, error) {
	return d.readUint32WithLength(offset, 3)
}
func (d Dat) readUint32WithLength(offset, len int64) (uint32, error) {
	var data uint32
	err := d.readWithSize(offset, &data, len, binary.LittleEndian)
	if err != nil {
		return 0, err
	}
	return data, nil

}
func (d Dat) readString(offset int64) (string, int64, error) {
	var data string
	r, err := d.fs.Seek(offset, 0)
	if err != nil {
		return data, r, err
	}
	reader := bufio.NewReader(d.fs)
	b, err := reader.ReadBytes(byte(0))
	if err != nil {
		return data, r, err
	}
	readed := len(b)
	b = b[:len(b)-1]
	b, err = simplifiedchinese.GBK.NewDecoder().Bytes(b)
	if err != nil {
		return data, r, err
	}
	data = string(b)
	return data, r + int64(readed), nil
}
func (d Dat) readWithSize(offset int64, data any, size int64, order binary.ByteOrder) error {
	l := int64(reflect.TypeOf(data).Elem().Size())
	tmp := make([]byte, size, l)
	_, err := d.fs.ReadAt(tmp, offset)
	if err != nil && err != io.EOF {
		return err
	}
	if data == nil {
		return fmt.Errorf("data nil")
	}
	for size < l {
		tmp = append(tmp, 0)
		size++
	}
	return binary.Read(bytes.NewReader(tmp), order, data)
}
func (d Dat) readAddress(offset int64) (country, area string, e error) {
	mode, err := d.readByte(offset + 4)
	if err != nil {
		return "", "", err
	}
	var curPointer int64
	switch mode {
	case 1:
		// 模式1
		countryOffset, err := d.readUint32WithLength3(offset + 5)
		if err != nil {
			return "", "", err
		}
		mode, err = d.readByte(int64(countryOffset))
		if err != nil {
			return "", "", err
		}

		if mode == 2 {
			countryOffsetOffset, err := d.readUint32WithLength3(int64(countryOffset) + 1)
			if err != nil {
				return "", "", err
			}
			country, _, err = d.readString(int64(countryOffsetOffset))
			if err != nil {
				return "", "", err
			}
			curPointer = int64(countryOffset) + 4
		} else {
			country, curPointer, err = d.readString(int64(countryOffset))
			if err != nil {
				return "", "", err
			}
		}
		area, err = d.readArea(curPointer)
		if err != nil {
			return "", "", err
		}
		return
	case 2:
		countryOffset, err := d.readUint32WithLength3(offset + 5)
		if err != nil {
			return "", "", err
		}
		country, _, err = d.readString(int64(countryOffset))
		if err != nil {
			return "", "", err
		}
		area, err = d.readArea(offset + 8)
		if err != nil {
			return "", "", err
		}
	default:
		country, curPointer, err = d.readString(offset + 4)
		if err != nil {
			return "", "", err
		}
		area, err = d.readArea(curPointer)
		if err != nil {
			return "", "", err
		}
	}
	return
}
func (d Dat) readArea(offset int64) (area string, err error) {
	mode, err := d.readByte(offset)
	if err != nil {
		return "", err
	}
	if mode == 1 || mode == 2 {
		var areaOffset uint32
		areaOffset, err = d.readUint32WithLength3(offset + 1)
		if err != nil {
			return "", err
		}
		if areaOffset == 0 {
			return "unknown area", nil
		}
		area, _, err = d.readString(int64(areaOffset))
		if err != nil {
			return "", err
		}
		return
	} else {
		area, _, err = d.readString(offset)
		if err != nil {
			return "", err
		}
		return
	}
}
func (d Dat) searchIndex(ip uint32) (uint32, error) {
	start, err := d.readUint32(0)
	if err != nil {
		return 0, err
	}
	end, err := d.readUint32(4)
	if err != nil {
		return 0, err
	}
	for {
		mid := getMidIndex(start, end)
		mip, err := d.readUint32(int64(mid))
		if err != nil {
			return 0, err
		}
		if end-start == 7 {
			endIpOffset, err := d.readUint32WithLength3(int64(mid) + 4)
			if err != nil {
				return 0, err
			}
			endIp, err := d.readUint32(int64(endIpOffset))
			if err != nil {
				return 0, err
			}
			if ip < endIp {
				return mid, nil
			} else {
				return 0, fmt.Errorf("not found ip")
			}
		}
		if ip > mip {
			start = mid
		} else if ip < mip {
			end = mid
		} else {
			return mid, nil
		}
	}
}

func getMidIndex(start, end uint32) uint32 {
	return start + (((end-start)/7)>>1)*7
}

func Uint32ToIp(ip uint32) net.IP {
	var b [4]byte
	b[0] = byte(ip & 0xff)
	b[1] = byte((ip >> 8) & 0xff)
	b[2] = byte((ip >> 16) & 0xff)
	b[3] = byte((ip >> 24) & 0xff)
	return net.IPv4(b[3], b[2], b[1], b[0])
}

func IpToUint32(ip net.IP) (u uint32) {
	ipStr := ip.String()
	ipStrS := strings.Split(ipStr, ".")
	u = 0
	tmp, _ := strconv.Atoi(ipStrS[3])
	u |= uint32(tmp) & 0xff
	tmp, _ = strconv.Atoi(ipStrS[2])
	u |= uint32(tmp) << 8 & 0xff00
	tmp, _ = strconv.Atoi(ipStrS[1])
	u |= uint32(tmp) << 16 & 0xff0000
	tmp, _ = strconv.Atoi(ipStrS[0])
	u |= uint32(tmp) << 24 & 0xff000000
	return u
}
