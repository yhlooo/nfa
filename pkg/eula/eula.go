package eula

import (
	"bufio"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/yhlooo/nfa/pkg/i18n"
)

//go:embed eula.md
var eulaFS embed.FS

const (
	eulaFileName      = "eula.md"
	signatureFileName = "eula_signed.sha256sum"
)

// Content 返回内嵌的 EULA 协议内容
func Content() string {
	data, err := eulaFS.ReadFile("eula.md")
	if err != nil {
		return ""
	}
	return string(data)
}

// SHA256 返回当前内嵌 EULA 内容的 SHA256 哈希值
func SHA256() string {
	data := Content()
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// Check 执行 EULA 协议确认检查
func Check(ctx context.Context, dataRoot string) error {
	// 每次启动时写入最新的 eula.md 到数据目录
	if err := writeToFile(dataRoot); err != nil {
		return fmt.Errorf("write eula file: %w", err)
	}

	currentSHA256 := SHA256()
	sigPath := filepath.Join(dataRoot, signatureFileName)

	// 检查签署状态
	signedSHA256, err := readSignedSHA256(sigPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read eula signature: %w", err)
	}

	if os.IsNotExist(err) {
		// 未签署过，首次启动
		return promptAndSign(ctx, sigPath, currentSHA256, false)
	}

	if signedSHA256 == currentSHA256 {
		// 已签署且版本一致，静默通过
		return nil
	}

	// SHA256 不匹配，协议已更新
	return promptAndSign(ctx, sigPath, currentSHA256, true)
}

// promptAndSign 展示 EULA 并询问用户是否同意
func promptAndSign(ctx context.Context, sigPath, sha256 string, isUpdate bool) error {
	content := Content()

	// 使用 glamour 渲染 Markdown 并打印
	rendered, err := glamour.Render(content, "auto")
	if err != nil {
		// 渲染失败时回退到原始文本
		fmt.Println(strings.Repeat("-", 72))
		fmt.Print(content)
		fmt.Println(strings.Repeat("-", 72))
		fmt.Println()
	} else {
		fmt.Print(rendered)
		fmt.Println()
	}

	if isUpdate {
		fmt.Println(i18n.TContext(ctx, MsgEULAUpdated))
		fmt.Println()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(i18n.TContext(ctx, MsgEULAAgreePrompt))
		if !scanner.Scan() {
			fmt.Println()
			fmt.Println(i18n.TContext(ctx, MsgEULADeclined))
			os.Exit(1)
		}

		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		switch answer {
		case "y", "yes":
			if err := writeSignedSHA256(sigPath, sha256); err != nil {
				return fmt.Errorf("write eula signature: %w", err)
			}
			return nil
		case "n", "no":
			fmt.Println(i18n.TContext(ctx, MsgEULADeclined))
			os.Exit(1)
			return nil
		default:
			fmt.Println(i18n.TContext(ctx, MsgEULAInvalidInput))
		}
	}
}

// writeToFile 将内嵌的 eula.md 写入数据目录
func writeToFile(dataRoot string) error {
	content := Content()
	path := filepath.Join(dataRoot, eulaFileName)
	return os.WriteFile(path, []byte(content), 0o644)
}

// readSignedSHA256 读取已签署的 SHA256 值
func readSignedSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// writeSignedSHA256 将 SHA256 值写入签名文件
func writeSignedSHA256(path, sha256 string) error {
	return os.WriteFile(path, []byte(sha256+"\n"), 0o644)
}
