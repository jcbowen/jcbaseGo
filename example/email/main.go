package main

import (
	"fmt"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/mailer"
)

func main() {
	// 创建一个新的Email实例
	e := mailer.New(&jcbaseGo.MailerStruct{
		Host:     "smtp.qiye.aliyun.com",
		Port:     "465",
		Username: "",
		Password: "",
		From:     "",
		UseTLS:   true,
	})

	// 添加收件人
	e.AddRecipient("563121793@qq.com")

	// 设置邮件主题
	e.SetSubject("测试邮件2")

	// Base64编码的图片数据（替换为你实际的Base64图片数据）
	base64Image := "iVBORw0KGgoAAAANSUhEUgAAAPAAAAA8CAIAAADXHaAKAAASBElEQVR4nOyceXAbZZr/W90tqVu3ZOu0fJ+x4yvYcRwHcgJJyGSGmQkDgV9mwo+FhQFqdmemdv+bqq2t2qplqandYYepZZiFSYblJkxICIGckDiJr/i+JVuWLFlX62y1pD62HAnZceJgO7Zb8fan/Eer+7X01auvnn7e531fwa6JNoCDY60A3nrKAegdgJ4NMRwcdws834XZntYB9tXSw8FxV8xr6Nlw5uZgkUXZb0GGXvKzc3AsjSUnvbcxdMKmC3lGztwcy8iyjNzmjdCzDcqZm2OFWPbyw4JSDs7cHMvIYk28KAstOofmzM2xBFbUxLNZtKHne9XFmpvz99pmCbnEsvjhrgw9m8Wae4WCNxmJR1y4xCDlwbeZM+JYUdgy8Wx4Kz31vbSsfxHvk2E8/S53l4OKkgAACOWIUInSJB3FIiRBgjDIMICiUKXbaFyCDI47s+If7uJZcUPPZslD2vm6wN5s8fQ6laWZunojhNz+bsNQDDbsNp8YzNlVqK7mpvTvimX/BJedVTX0HO6ydyxnRiEhnLUldyH/EsUixw/8z47ffU+1Tr20F/2/Sfo7eA5sGno2i+04OkaOH72c//QDC++7669d8Y14Sg5UGppylipz7XM3heF0GOUv26DwLpnTF9/ZrYFem2pjwXyNb+1Z3BGyXhyreaGh4z8uB8awsierl0P1Pc9dzmukg4PnkC4R+s7c2u/RqcDI779SbsgDQF7UFSTsPjIUJab8cR8OCiBQAKNZqsp/OZBq3/WP76kNfEgICaTCOB7P3l6gKFSt+vtgk2WZk0tDB88hXSL0nbk1fgu1stJf7W17/i31/aUMRUvL9AxJB/ps8DoDGSJAIZ9w+FLth179HKAZOieXcIfiiFhgkAy3BbILK2775Pc6yzuZfM91zr0RoecjbA9+cfgjRWOZpFQPIXxEp5j6oosvF9n+2l73xv+XFGoAABh85WTEhgmUIliCCLVyHsSLY3hGY5GidkGjyTT8RFd0+8US3i9DMzyQtzJyFs29EaHnQ6yXlh+qbf/3yxtKRLg5iA9RBQ268dMjZf+wL+Hmod+ewtrMhh/cR0diUVcw0e90jFygm5fRPTrAnob7gJb2dY3isYELgyEszNAMAAB8lM9QDCIV1uypWgGNi+PejtAJRo/19R29LstVqKt0uDMskAqr/nYjAAC9f2pzdjrk+UoYhUVqCRYTO77oRvSK0l/uEWRI2Fa92izLrcY15h5tMWfmZGRXGoUiwexL/il/x8mupoOb+EL+3b/QklkLhk5B4vGeN1trXmoEAGDirAkUQN5+Z8SFi/VSHsSjCNLZYa95aVNGuWb2f6Vh4LwbVi5Haj3WjkiQ9bvK52sQDUc/+7dT9Y9uyKnKXiEN38m9nXLMofet9vJDtQAAxALRwfe6G3+zI2T1IyqRb9SjLM60dTrKDlbNcfN3OiAN7c5KWn/1o1Z9sTbl1Hg0Pt5hiUfJkqYiCIYSJ4ViYcnmoiN//+4zf/iptmhuP68OaydCOzsmcUcob08JAAB9b7f7TRhfIqh9ubH3v9uFCgQW8QUSYfaOArbkOQB9Go4vF0jIE+o63bv5iYbEw+b3r0322/kInwgSUTz65Cs/AaHkUjCn2XX2jQvYpO+5N59OnVxN1siSNIZmxk+PJNxMeCP2q9bC768TyIRD7/eUPl7lG/Wy6+b0rJYsnJFrplSmce7NiyqDQqISawrUKqMS4kN95wZSLb1WrGZ3JSpDWz/tYEXqGjH00Ac9JQcqE8d9f+6Q5SisF8eUJZmQEOp9q11Ta2DXzfc6cSIuvTGMHm4eGWsfj+IxhUGB+3FUjuqKtR6rN9XSa8Vya3JUWYrh5hFWpM7k0GH3lGd8BPe6Y+FgnIjQJMkwdOISjwfyQB4PgiGYD/EFfJFIrjNqy6pBOC1S8JifiPkJeYEycRyyBdRVOt+oF58KBa0BkVqUv7eEbY3fDUnFYYjN+sAdKKjL7zs/wBfCVz5s0RZpouEoGSNpinaOOnkgWNpUlGoZj8Y7T/U0/Kj+s1dPYZOY0qBcZalJR7pG+js/fpuMxQAAgPh8WIiCMMzjJavlDMMwFEWTcZoiaYqiKQoAAJnuQuPTv+CB7Mf4oQ96yp6sSRyPHh+475dNtotjmlq9s8MeD0Xv+7smtgXeCYvdZLINOb12t8/J4wECvrC2bFOOLj9ToWVb2gwZ2arOU922/snybWVklBptMRsrshiKiQSJhh/XF9bnp1p6Jrzl28q6z/SWby9DpOjqS00aevj8SYFY1vCzwyJlBsS/qb4Yw0MAwBOIxKkzVDw2fu3i0LmTHvNQZmHZqmu+iZAtgKrFfHEytjk77DACq6v19isWkVaiebiIB6XLJNYc7G7rgLmrLL9qW93uRNSIxgmv36XPzD7feipHX1BoLGVb4wzbn3ngy9fPucxuMk4FXEGnyVXaVLzr+e2pqDedb9iwKB7l8YDyrWUtn7ShUmT1dSYNzTAMH0GkmmSJyjs2PNnd5hkbjgYDNE2BILTp8MsyfbJkA/EFhqr6oXMnA3Yr64Ye/rCn9uXGxLH9yoRuozEejo1+2r/+mbqeP7bqG9N0peg3189MOi2P7niKDye/ijweDxGgBvW04O31e05d/kQpVank6bJ6OxIg+EKYjJFl95fsfnnXbWdPrL02VIp6rFjIEzaWZ7EhM2VoigIFQgAAAnbrwJfHvBYTCEEKYz4qV3otJpqmGJqe/W8CVHyjkB5kRXQK63mzbqMR+DZITLVYaYrJqNBIcxTNvzlT+Td17MqbD5tzfMDc9f/2PZ9y86083Pj9tv7mNDG00+Tq/qp3y1Ob58wOem0YKkNTkdgz4c2pMjpNLq8N2/3Sg6xITRqaCGAStW7wzPGxK+chPr9k+yPZGxrDnqm2d/8IwnDF3gPyrJsWP0xn2CBIxghWRKdwdTlS4ZmO07gzXHaw2nrRjGaKERWaWaVjV95ticaJzy5+sKthn5A/c0eORPFJ1wQP4GXr8hMu5/HA9UUbxu2jufpCVvVO29TSPbHz2W2pM+PXLT1n+z0TnpLGYv+U/+GXdt1IRCmGZkAQlKllIU9YZVzt4WCCaUPjmJuMxRz9nQAAqIvL1z/ymFAiwyZMre/8FywU1j/5vEw/s8OU8GNjLRddQ30MTUsy2XSMq9M+e/+VvdmCqFB3z1RGuQYb8mjr03RX7BeXj8EQXJg9k6r1m7u6hlozFGqSIs+3nWqq2VGaux4AAESAhnCW74EUSQ03j256rD7x0Gf3Nb/fQsVJSYaksC4v4ArifjxxafCbIblWRoSjFElJM1lbKjNtaJ91bDroglDpg/tz6+8HACDscba/9yYsRBoO/VykmrnrmZvPDl84RZOkIiundOe+vIatbOkGAMDRYqt8Ziap8A64NBsMPB7PO+DChjyJOfB0Y9Q6aLIO1VXMFF6uD14bnRjwhzCZWL5ny48oijxy4vUMuTpR5SCiOKt6gYGLQ9W71yeOvVbsk3/+a25tjiY/U1uocZnd5vbxom/3DRHhqMeKVWwvI0JRc/sYW4KnDa0uKtdXbNBX1GpKKm5U0fG2d99gaKruiRk3kzGi85OjruE+ZXZ++cM/lOrYSflnYBiIP1MxjAWjNEmDMBS0+uUFKr8ZA/ns1xNvZXxypMBYYtQkbyw3qnVTP9p1iKTIT8+9Y7IOFhhLN1VubetrfnjzD6bzKIbGiZAIYS3ghX04KktW374+ckmmlRnXGTC7f7xrAhbAYSxcsTM5g0jFKXVOhq3fLlGKUqs7Vh/wxnpWUfWjTyXcDABA58dHcMxTuf+gVGtInInh4atvv+Ya7ivYvLPh0IvsuxkAnNftmetnEh7ziUFZriJo9es3Gi1nRgv2sVx7mQ+aoWmaSg31uoZbt9ftmY4rELx/2+PtA1cZgNGrswV8YaJBNB6NRCMsCga/LXpGgkTYhxtK9ab2caVBMXJlFPfh2kINIklKpeK0KluFShE8EIniMdYEz3k8duW82zSY1/CAtiy5WDtO4C1Hfx+cmqzYe6BkxyMALy3Kuu4uh7pmZh2cp88Z9U2PUC1nTUI5YticptW66RELERYj4kSpFBGgEJQcl/NhgVGTY5sad/ucWZqk/kDIx4cFLKqN3/j5nhsDLRzmwzRFy7Wy1mPtW392PypFGx9vSLVU6OXdX/bCAhiCIb6QtSnkmwwdnJocOndCkZVbunN/4gwZI1r/8oeg077+kceyN3xb7u1pb3v3ja5jf7G0fM2G5mlAPpSaMfGbsPBkUCATxvwEFSWFchbq+QshEPYL+UgsHkuYmIhFEKFodoM4GbdOjQ2P9+VnFSfOeP0uCSplSS+QWPWVOMjIUWXkqCxdEx6LhwgRlu4JsVKkypopZeRWZ8fwWCQYARggTsTZEjzzTaIpsvPYURDmVz/6VHJCm2Guf/iW325d/8gBY+2mRLOgw9Z57CgfQal4fLKnLefGIHL1YSgmdTz0fnfh99eF7UGapL2D7vpfsyPpOwmGfYgQhUCIoikIhBAB4vW7Zjdweu0kRRq1uYmoTMQiqFAEsrq4gAjNVGZ3PrvN1Drmc/hq9lTpS3R85KYiOipDaYrGfRG5VsbKwtEEMy88dPZEyOUo3/1DVJGRODN84ZTbNFSyba+xtjHVzNR8DgShpmd/XfbgfjYEJ4l4ksN/fCrkN2N9RzqQDJGyJBNgGIlRxqKwO0DECESAihBxOBJKVJpFqOTzSx+PTPRPeSbHJkecmIPH4zVUPpBo3zFwZVv9bnY164t1A18PpR4W1OVt2FeTU5WdcPPp/zyDTWKpqyIFigciMTyG+1nL+5OGnhrsHrt6Iauq3lCZLIT5bGOmS19pissLtuxKtfaYBu297YaqOqFEZrp8VqzKZEl28lZIx+kr/3RWWZJR+L11nj5nYNynqUm7DSYpGIYWo1KNSu/yJtdGb6nZWV1SZ3dZu4ZbXZjjJw8dfuyhw4k1d+bJYSEfyZCzs+8jxbqtpX3nBkaumuacxyax4/96EpWh1z/vTp10jXlQKeKxYeo81owxnXJEfJ6e4+9KMrUVe3+cujD6zVcMwzAAcO3Pr8FCBFWoQAi2tF1GZIqyB/d7LaNEwLd+32Ns6S47WPXlM59MH4E8AotQMUqskwAMo0vXxRsAAAj5SCDsz8sq6hnpSE2sGNQ5ifUbs3G4bWO2ke31e9iQOZd9v9r9p58fufTOlZwqo65IAwA8a5/NPxXQFWs1eZmRAHHi1VN8VOCz+4gQAcEQBINBd5CMkjAbQ8PpCH394yNxIqLKK/KOjwSd9mgowFAUdGMC1j3STwT9Ptv4eMs35ivnRarM+oPPwUI06LBNDxTyildfcQKxTlr9QgPAA9YdrEYzRGQkThGk/Zr11i2D6YNCluH02rO1+VjAgwU88zXrHm4btvSxnmykEIgET77yE3VehteKjXdZR66ZwhhevrU0uyIr6AkRYYKMU3KtzGPxPPTCjtEWk7ZAg9l9rLg5GaETww5L6yVL66XUBRCCAABAZApxhgaRKRCZQqTMVBevgwUIAAB++wTEF6SybVbQbDBUPbex7+2O0serJpstSKZY38DaZuOFIBXJQngAAIA9TT88+c1HP9j+hPjmCoY34O4cvJZnKK4svo89mbdBmikprM93W7yWrokHDjX1nR+w9k2KlWJwOhiHHnphR2JEePK3pxsO1A1eGt52mLVxeXKTLBWL4pibCPijoUA0HIxN/4VikTBJRGJ4OBoO0iSZcLm2tLJ01/6rb/8OlSs3HnqRLd0pTMcHOl+/Ks9XSnMUG36xGWKvAroQrg9eU0hVeYYiF+a42n0xsWQUBEEBX+j2OU3WwZ/uf1ElYy0BvTPdX/aa28dcY57cmuzRa2Z9ibZsS0lJ08xdOuQNO4ancquz5xRAVpOF7vomAj6fbdw51Ovo7eCBIEXGS7bvLWjatfIK1xqtfZfjZDRToZWK5SAPxImwP4SN20ctdtPupkeLstexLfDeZtE/YxB02ke/Ph0LB2t+fHj2NhaOhUNS8UnnhD/sI8lYnIxHY4RCqirOLUcELOxZWmOsnd/l4OBYOz9jwMGRgDM0x5qCMzTHmoIzNMeagjM0x5qCMzTHmoIzNMeagjM0x5qCMzTHmoIzNMeagjM0x5qCMzTHmuJ/AwAA//9MftOetI2yMAAAAABJRU5ErkJggg=="

	// 设置HTML正文，包含内嵌图片
	htmlBody := `<html><body><h1>您的验证码如下：</h1><img src="cid:image1"></body></html>`
	e.SetBody(htmlBody, true)

	// 添加内嵌图片
	e.AddInlineImage("image1", base64Image)

	// 发送邮件
	err := e.Send()
	if err != nil {
		fmt.Println("发送邮件出错:", err)
	} else {
		fmt.Println("邮件发送成功!")
	}
}
