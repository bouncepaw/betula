package types

import "testing"

func TestCleanerLinkParts(t *testing.T) {
	check := func(url string, expectedLeft string, expectedRight string) {
		left, right := CleanerLinkParts(url)
		if left != expectedLeft {
			t.Errorf("Wrong left part for `%s`: expected `%s`, got `%s`", url, expectedLeft, left)
		}
		if right != expectedRight {
			t.Errorf("Wrong right part for `%s`: expected `%s`, got `%s`", url, expectedRight, right)
		}
	}

	check("gopher://foo.bar/baz", "gopher://foo.bar", "/baz")
	check("https://example.com/", "example.com", "")
	check("http://xn--d1ahgkh6g.xn--90aczn5ei/%F0%9F%96%A4", "юникод.любовь", "/🖤")
	check("http://юникод.любовь/🖤", "юникод.любовь", "/🖤")
	check("http://example.com/?query=param#a/b", "example.com", "?query=param#a/b")
	check("mailto:user@example.com", "mailto:user@example.com", "")
	check("tel:+55551234567", "tel:+55551234567", "")
}
