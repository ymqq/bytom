package compiler

import (
	"fmt"
	"github.com/bytom/exp/ivy/compiler/semver"
)

const IVY_VERSION string = "1.0.0"

func parseVersion(p *parser) bool {
	if peekKeyword(p) == "pragma" {
		consumeKeyword(p, "pragma")
		if peekKeyword(p) == "ivy" {
			consumeKeyword(p, "ivy")
			strliteral, newOffset := scanVersionStr(p.buf, p.pos)
			if newOffset < 0 {
				p.errorf("Invalid version character format!")
			}
			p.pos = newOffset

			//After removing the quotes is the version info
			version := strliteral[1 : len(strliteral)-1]
			fmt.Println("version info:", string(version))
			if ok := checkVersion(string(version)); ok {
				return true
			}
			return false
		}
		return false
	}

	//when contract is not contain the version info, return true
	return true
}

func checkVersion(version string) bool {
	c, err := semver.NewConstraint(version)
	if err != nil {
		panic(err)
	}

	v, err := semver.NewVersion(IVY_VERSION)
	if err != nil {
		panic(err)
	}

	return c.Check(v)
}

func scanVersionStr(buf []byte, offset int) ([]byte, int) {
	offset = skipWsAndComments(buf, offset)

	//the talbe of ascii code for double quote and single quote:
	//  \" -- 0x22/34
	//  \' -- 0x27/37
	if offset >= len(buf) || !(buf[offset] == '\'' || buf[offset] == 34) {
		return nil, -1
	}

	for i := offset + 1; i < len(buf); i++ {
		if (buf[offset] == '\'' && buf[i] == '\'') || (buf[offset] == 34 && buf[i] == 34) {
			return buf[offset : i+1], i + 1
		}
		if buf[i] == '\\' {
			i++
		}
	}

	panic(parseErr(buf, offset, "unterminated version string literal"))
}