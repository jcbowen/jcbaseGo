package mailer

import (
	"crypto/tls"
	"fmt"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/tlsconfig"
	"log"
	"net/smtp"
	"strings"
)

// Email 结构体定义了邮件的基本属性
type Email struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	From         string
	To           []string
	Subject      string
	Body         string
	IsHTML       bool          // 是否为HTML正文
	InlineImages []InlineImage // 内嵌图片
	UseTLS       bool          // 是否使用TLS
	CertFile     string        // 证书文件路径
	KeyFile      string        // 私钥文件路径
	CAFile       string        // CA证书文件路径（可选）
}

// InlineImage 结构体定义了内嵌图片的基本属性
type InlineImage struct {
	CID  string // 内容ID，用于在HTML中引用
	Data string // Base64编码的图片数据
}

// New 创建一个新的Email实例
func New(conf *jcbaseGo.MailerStruct) *Email {
	return &Email{
		SMTPHost: conf.Host,
		SMTPPort: conf.Port,
		SMTPUser: conf.Username,
		SMTPPass: conf.Password,
		From:     conf.From,
		UseTLS:   conf.UseTLS,
		CertFile: conf.CertFile,
		KeyFile:  conf.KeyFile,
		CAFile:   conf.CAFile,
	}
}

// AddRecipient 添加收件人地址
func (e *Email) AddRecipient(to string) {
	e.To = append(e.To, to)
}

// SetSubject 设置邮件主题
func (e *Email) SetSubject(subject string) {
	e.Subject = subject
}

// SetBody 设置邮件正文
func (e *Email) SetBody(body string, isHTML bool) {
	e.Body = body
	e.IsHTML = isHTML
}

// AddInlineImage 添加内嵌图片
func (e *Email) AddInlineImage(cid, base64Data string) {
	e.InlineImages = append(e.InlineImages, InlineImage{CID: cid, Data: base64Data})
}

// Send 发送邮件
func (e *Email) Send() error {
	var err error
	var client *smtp.Client

	if e.UseTLS {
		// 使用TLS连接
		tlsConfig, err := tlsconfig.Get(e.CertFile, e.KeyFile, e.CAFile, e.SMTPHost)
		if err != nil {
			return fmt.Errorf("无法获取TLS配置: %v", err)
		}

		hostPort := e.SMTPHost + ":" + e.SMTPPort
		conn, err := tls.Dial("tcp", hostPort, tlsConfig)
		if err != nil {
			return fmt.Errorf("无法建立TLS连接: %v", err)
		}
		defer func(conn *tls.Conn) {
			_ = conn.Close()
		}(conn)

		client, err = smtp.NewClient(conn, e.SMTPHost)
		if err != nil {
			return fmt.Errorf("创建SMTP客户端失败: %v", err)
		}
	} else {
		// 使用普通连接
		addr := e.SMTPHost + ":" + e.SMTPPort
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("无法建立SMTP连接: %v", err)
		}
		defer func(client *smtp.Client) {
			err := client.Close()
			if err != nil {
				log.Println("关闭SMTP连接失败:", err)
			}
		}(client)
	}

	auth := smtp.PlainAuth("", e.SMTPUser, e.SMTPPass, e.SMTPHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	if err = client.Mail(e.From); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	for _, to := range e.To {
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败: %v", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取SMTP数据写入器失败: %v", err)
	}

	msg := e.buildMessage()
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("发送邮件内容失败: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("关闭SMTP数据写入器失败: %v", err)
	}

	err = client.Quit()
	if err != nil {
		return fmt.Errorf("关闭SMTP客户端失败: %v", err)
	}

	return nil
}

// buildMessage 构建邮件消息
func (e *Email) buildMessage() string {
	boundary := "boundary"
	var msgBuilder strings.Builder

	// 构建邮件头部
	msgBuilder.WriteString("From: " + e.From + "\n")
	msgBuilder.WriteString("To: " + strings.Join(e.To, ", ") + "\n")
	msgBuilder.WriteString("Subject: " + e.Subject + "\n")
	msgBuilder.WriteString("MIME-Version: 1.0\n")

	if len(e.InlineImages) == 0 {
		// 如果没有内嵌图片，直接构建邮件正文
		if e.IsHTML {
			msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\n")
		} else {
			msgBuilder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\n")
		}
		msgBuilder.WriteString("Content-Transfer-Encoding: 7bit\n\n")
		msgBuilder.WriteString(e.Body)
	} else {
		// 如果有内嵌图片，构建multipart邮件
		msgBuilder.WriteString("Content-Type: multipart/related; boundary=" + boundary + "\n")

		// 添加邮件正文部分
		msgBuilder.WriteString("\n--" + boundary + "\n")
		if e.IsHTML {
			msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\n")
		} else {
			msgBuilder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\n")
		}
		msgBuilder.WriteString("Content-Transfer-Encoding: 7bit\n\n")
		msgBuilder.WriteString(e.Body + "\n")

		// 添加内嵌图片
		for _, image := range e.InlineImages {
			msgBuilder.WriteString("\n--" + boundary + "\n")
			msgBuilder.WriteString("Content-Type: image/jpeg; name=\"" + image.CID + "\"\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: base64\n")
			msgBuilder.WriteString("Content-Disposition: inline; filename=\"" + image.CID + "\"\n")
			msgBuilder.WriteString("Content-ID: <" + image.CID + ">\n\n")
			msgBuilder.WriteString(splitBase64(image.Data) + "\n")
		}

		// 结束boundary
		msgBuilder.WriteString("--" + boundary + "--")
	}

	return msgBuilder.String()
}

// splitBase64 将Base64编码的数据分行，以符合RFC 2045标准
func splitBase64(encoded string) string {
	var result []string
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		result = append(result, encoded[i:end])
	}
	return strings.Join(result, "\n")
}
